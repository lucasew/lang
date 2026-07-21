package detector

import (
	"archive/zip"
	"bufio"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"golang.org/x/text/unicode/norm"
)

// NGramDetector ports org.languagetool.language.identifier.detector.NGramDetector.
// When loaded from a ZIP (NewNGramDetectorFromZip), uses Java vocab + Kneser-Ney
// bigram model. Otherwise CharNGramDetector is a test-only fallback surface.
type NGramDetector struct {
	*CharNGramDetector
	MaxLength int

	// ZIP model fields (Java NGramDetector)
	vocab           map[string]int
	codes           [][]string // each: name, 2-code or NULL, 3-code, flag…
	knpBigramProbs  []map[string]float64
	thresholdsStart int
	thresholds      [][]float64
	zipLoaded       bool
}

const ngramEpsilon = 1e-4

var (
	ngramDigits   = regexp.MustCompile(`\d+`)
	ngramKorean   = regexp.MustCompile(`[\x{ac00}-\x{d7a3}]`)
	ngramJapanese = regexp.MustCompile(`[\x{3040}-\x{30ff}]`)
	ngramChinese  = regexp.MustCompile(`[\x{4e00}-\x{9FFF}]`)
	ngramKhmer    = regexp.MustCompile(`[\x{1780}-\x{17FF}]`)
	ngramTagalog  = regexp.MustCompile(`[\x{1700}-\x{171F}]`)
	ngramArmenian = regexp.MustCompile(`[\x{0530}-\x{058F}]`)
	ngramGreek    = regexp.MustCompile(`[\x{0370}-\x{03FF}]`)
	ngramTamil    = regexp.MustCompile(`[\x{0B80}-\x{0BFF}]`)
	// Java Pattern.compile("\\s+") — ASCII whitespace without UNICODE_CHARACTER_CLASS
	ngramWhitespace = regexp.MustCompile(`[ \t\n\x0B\f\r]+`)
)

func NewNGramDetector(maxLength int) *NGramDetector {
	if maxLength <= 0 {
		maxLength = 1000
	}
	return &NGramDetector{
		CharNGramDetector: NewCharNGramDetector(3),
		MaxLength:         maxLength,
	}
}

// NewNGramDetectorFromZip ports NGramDetector(File sourceModelZip, int maxLength).
func NewNGramDetectorFromZip(zipPath string, maxLength int) (*NGramDetector, error) {
	if maxLength <= 0 {
		maxLength = 50 // Java DefaultLanguageIdentifier.enableNgrams uses 50
	}
	d := NewNGramDetector(maxLength)
	if err := d.loadZip(zipPath); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *NGramDetector) loadZip(zipPath string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()
	byName := map[string]*zip.File{}
	for _, f := range zr.File {
		byName[f.Name] = f
		// also bare basename
		if i := strings.LastIndex(f.Name, "/"); i >= 0 {
			byName[f.Name[i+1:]] = f
		}
	}
	readEntry := func(name string) ([]string, error) {
		f, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("ngram zip missing entry: %s", name)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		var lines []string
		sc := bufio.NewScanner(rc)
		// large lines possible in transition matrices
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 16*1024*1024)
		for sc.Scan() {
			lines = append(lines, sc.Text())
		}
		if err := sc.Err(); err != nil {
			return nil, err
		}
		return lines, nil
	}

	// iso_codes.tsv — Line: {Name}\t{2-code or NULL}\t{3-code}\t{flag}
	codeLines, err := readEntry("iso_codes.tsv")
	if err != nil {
		return err
	}
	var codes [][]string
	for _, line := range codeLines {
		values := strings.Split(line, "\t")
		// Java: values[3].equals("1")
		if len(values) > 3 && values[3] == "1" {
			codes = append(codes, values)
		}
	}
	d.codes = codes

	// vocab.txt — token\t… first column is token
	vocabLines, err := readEntry("vocab.txt")
	if err != nil {
		return err
	}
	d.vocab = map[string]int{}
	for i, line := range vocabLines {
		tok := line
		if parts := strings.Split(line, "\t"); len(parts) > 0 {
			tok = tools.JavaStringTrim(parts[0])
		}
		d.vocab[tok] = i
	}

	// thresholds.txt
	thLines, err := readEntry("thresholds.txt")
	if err != nil {
		return err
	}
	if len(thLines) == 0 {
		return fmt.Errorf("ngram zip thresholds.txt empty")
	}
	start, err := strconv.Atoi(tools.JavaStringTrim(thLines[0]))
	if err != nil {
		return fmt.Errorf("thresholds start: %w", err)
	}
	d.thresholdsStart = start
	for _, line := range thLines[1:] {
		parts := strings.Split(tools.JavaStringTrim(line), " ")
		row := make([]float64, 0, len(parts))
		for _, p := range parts {
			if p == "" {
				continue
			}
			v, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return err
			}
			row = append(row, v)
		}
		d.thresholds = append(d.thresholds, row)
	}

	// transition matrices 00.txt … per language index
	d.knpBigramProbs = make([]map[string]float64, len(d.codes))
	for i := 0; i < len(d.codes); i++ {
		name := fmt.Sprintf("%02d.txt", i)
		lines, err := readEntry(name)
		if err != nil {
			return err
		}
		d.knpBigramProbs[i] = loadNGramDict(lines)
	}
	d.zipLoaded = true
	return nil
}

func loadNGramDict(lines []string) map[string]float64 {
	tm := map[string]float64{}
	for _, line := range lines {
		parts := strings.Fields(tools.JavaStringTrim(line))
		if len(parts) < 2 {
			continue
		}
		key := strings.Join(parts[:len(parts)-1], "_")
		v, err := strconv.ParseFloat(parts[len(parts)-1], 64)
		if err != nil {
			continue
		}
		tm[key] = v
	}
	return tm
}

// DetectLanguages scores text (char-ngram fallback or ZIP model without additional codes).
func (d *NGramDetector) DetectLanguages(text string) map[string]float64 {
	return d.DetectLanguagesAdditional(text, nil)
}

// DetectLanguagesAdditional ports detectLanguages(text, additionalLanguageCodes).
func (d *NGramDetector) DetectLanguagesAdditional(text string, additionalLanguageCodes []string) map[string]float64 {
	if d == nil {
		return nil
	}
	if d.zipLoaded {
		return d.detectLanguagesZip(text, additionalLanguageCodes)
	}
	// Char n-gram fallback (tests without ZIP)
	if d.MaxLength > 0 && javaStringLenDet(text) > d.MaxLength {
		text = javaSubstringDet(text, 0, d.MaxLength)
	}
	scores := map[string]float64{}
	if d.CharNGramDetector != nil {
		if s := d.CharNGramDetector.DetectLanguages(text); s != nil {
			scores = s
		}
	}
	if isDigitsOnlyText(text) {
		return map[string]float64{"zz": 1}
	}
	// script boosts for char-ngram path only (ZIP encode handles scripts via tags)
	boostScripts(scores, text)
	return scores
}

func (d *NGramDetector) detectLanguagesZip(text string, additional []string) map[string]float64 {
	// Java: List<Integer> enc = encode(text); … text.length() is UTF-16
	enc := d.encode(text)
	keys := ngramKeys(enc)
	finalProbs := make([]float64, len(d.codes))
	for i := range d.codes {
		var val float64
		tm := d.knpBigramProbs[i]
		for _, key := range keys {
			k := strconv.Itoa(key[0]) + "_" + strconv.Itoa(key[1])
			prob, ok := tm[k]
			if !ok {
				prob = ngramEpsilon
			}
			val += math.Log(prob)
		}
		finalProbs[i] = val
	}

	result := map[string]float64{}
	textLen := javaStringLenDet(text)
	if textLen >= d.thresholdsStart && len(d.thresholds) > 0 {
		argMax := 0
		for i := 1; i < len(finalProbs); i++ {
			if finalProbs[i] > finalProbs[argMax] {
				argMax = i
			}
		}
		thresholdIndex := minInt(textLen, d.MaxLength) - d.thresholdsStart
		if thresholdIndex < 0 {
			thresholdIndex = 0
		}
		if thresholdIndex >= len(d.thresholds) {
			thresholdIndex = len(d.thresholds) - 1
		}
		row := d.thresholds[thresholdIndex]
		if argMax < len(row) && finalProbs[argMax] < row[argMax] {
			// Java NoopLanguage.SHORT_CODE = "zz", score 100.0
			result["zz"] = 100.0
			return result
		}
	}

	// exp + normalize
	for i := range finalProbs {
		finalProbs[i] = math.Exp(finalProbs[i])
	}
	var tot float64
	for _, v := range finalProbs {
		tot += v
	}
	if tot > 0 {
		for i := range finalProbs {
			finalProbs[i] /= tot
		}
	}
	for i := range d.codes {
		langCode := d.codes[i][1]
		if langCode == "NULL" && len(d.codes[i]) > 2 {
			langCode = d.codes[i][2]
		}
		if canDetectLang(langCode, additional) {
			result[langCode] = finalProbs[i]
		}
	}
	return result
}

func canDetectLang(langCode string, additional []string) bool {
	// Defer full Languages.isLanguageSupported — accept all non-empty + additional.
	// Callers filter further via DefaultLanguageIdentifier.
	if langCode == "" {
		return false
	}
	_ = additional
	return true
}

func (d *NGramDetector) encode(text string) []int {
	result := []int{1} // start of sentence
	if d.MaxLength > 0 && javaStringLenDet(text) > d.MaxLength {
		text = javaSubstringDet(text, 0, d.MaxLength)
	}
	// Java: Normalizer.NFKC + toLowerCase()
	text = norm.NFKC.String(text)
	text = strings.ToLower(text)
	text = ngramDigits.ReplaceAllString(text, "<NUM>")
	text = ngramKorean.ReplaceAllString(text, "<KO>")
	text = ngramJapanese.ReplaceAllString(text, "<JA>")
	text = ngramChinese.ReplaceAllString(text, "<ZH>")
	text = ngramKhmer.ReplaceAllString(text, "<KM>")
	text = ngramTagalog.ReplaceAllString(text, "<TL>")
	text = ngramArmenian.ReplaceAllString(text, "<HY>")
	text = ngramGreek.ReplaceAllString(text, "<EL>")
	text = ngramTamil.ReplaceAllString(text, "<TA>")
	text = ngramWhitespace.ReplaceAllString(text, "▁")
	if len(text) == 0 {
		return result
	}
	text = "▁" + text
	// Java String indexing is UTF-16; use UTF-16 code units for substring greediness.
	u := utf16.Encode([]rune(text))
	cur := 0
	for cur < len(u) {
		tok := 0
		ci := 1
		for i := cur + 1; i <= len(u); i++ {
			sub := string(utf16.Decode(u[cur:i]))
			if maybeTok, ok := d.vocab[sub]; ok {
				tok = maybeTok
				ci = i - cur
			}
		}
		cur += ci
		result = append(result, tok)
	}
	return result
}

func ngramKeys(enc []int) [][2]int {
	var out [][2]int
	for i := 1; i < len(enc); i++ {
		out = append(out, [2]int{enc[i-1], enc[i]})
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func boostScripts(scores map[string]float64, text string) {
	var hasKo, hasJa, hasZh, hasKm, hasTl, hasHy, hasEl, hasTa bool
	for _, r := range text {
		switch {
		case r >= 0xac00 && r <= 0xd7a3:
			hasKo = true
		case r >= 0x3040 && r <= 0x30ff:
			hasJa = true
		case unicode.Is(unicode.Han, r):
			hasZh = true
		case r >= 0x1780 && r <= 0x17ff:
			hasKm = true
		case r >= 0x1700 && r <= 0x171f:
			hasTl = true
		case r >= 0x0530 && r <= 0x058f:
			hasHy = true
		case r >= 0x0370 && r <= 0x03ff:
			hasEl = true
		case r >= 0x0b80 && r <= 0x0bff:
			hasTa = true
		}
	}
	if hasKo {
		scores["ko"] += 2
	}
	if hasJa {
		scores["ja"] += 2
	}
	if hasZh {
		scores["zh"] += 2
	}
	if hasKm {
		scores["km"] += 2
	}
	if hasTl {
		scores["tl"] += 1.5
	}
	if hasHy {
		scores["hy"] += 1.5
	}
	if hasEl {
		scores["el"] += 1.5
	}
	if hasTa {
		scores["ta"] += 1.5
	}
}

// TopLanguage returns the best code or "" if empty.
func (d *NGramDetector) TopLanguage(text string) string {
	scores := d.DetectLanguages(text)
	var best string
	var bestV float64 = -1
	for k, v := range scores {
		if v > bestV {
			bestV = v
			best = k
		}
	}
	return best
}

func isDigitsOnlyText(text string) bool {
	s := tools.JavaStringTrim(text)
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) && !tools.CharacterIsWhitespace(r) {
			return false
		}
	}
	return true
}

func javaStringLenDet(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func javaSubstringDet(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}
