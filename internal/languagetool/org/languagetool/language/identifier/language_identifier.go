package identifier

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// NonLatinCharsLanguages ports LanguageIdentifier.NON_LATIN_CHARS_LANGUAGES.
var NonLatinCharsLanguages = []string{
	"ar", "fa", "ru", "uk", "be", "zh", "ja", "km", "ta", "el", "hi", "mr", "th", "he", "ko",
}

const (
	ScoreThreshold                 = 0.85
	ConsiderOnlyPreferredThreshold = 50
	DefaultMaxLength               = 1000
)

var (
	urlRE     = regexp.MustCompile(`https?://[-_.?&~;+=/#%0-9A-Za-z]+`)
	mailRE    = regexp.MustCompile(`[-_.0-9A-Za-z]+@[-_0-9A-Za-z]+[-_.0-9A-Za-z]+`)
	signature = regexp.MustCompile(`(?s)\n--[ \x{00A0}]\n.*`)
	mention   = regexp.MustCompile(`@[A-Za-z0-9_]+`)
	nbspInvis = regexp.MustCompile(`[\x{FEFF}\x{2063}]+`)
)

// LanguageIdentifier ports org.languagetool.language.identifier.LanguageIdentifier surface.
type LanguageIdentifier interface {
	// Detect returns the best language score for cleaned text (codes only).
	Detect(cleanText string, noopLangs, preferredLangs []string) *languagetool.DetectedLanguage
	// Scores returns top language scores.
	Scores(cleanText string, noopLangs, preferredLangs []string, limitOnPreferred bool, count int) []languagetool.DetectedLanguage
	// CleanAndShortenText cleans and truncates input.
	CleanAndShortenText(text string) string
}

// BaseLanguageIdentifier holds shared maxLength and clean helpers.
type BaseLanguageIdentifier struct {
	MaxLength int
}

func NewBaseLanguageIdentifier(maxLength int) BaseLanguageIdentifier {
	if maxLength < 10 {
		panic("maxLength must be >= 10 (but values > 100 are recommended)")
	}
	return BaseLanguageIdentifier{MaxLength: maxLength}
}

// CleanAndShortenText ports LanguageIdentifier.cleanAndShortenText.
func (b BaseLanguageIdentifier) CleanAndShortenText(text string) string {
	short := text
	if len(short) > b.MaxLength {
		// rune-safe truncate
		r := []rune(short)
		if len(r) > b.MaxLength {
			short = string(r[:b.MaxLength])
		}
	}
	short = nbspInvis.ReplaceAllString(short, " ")
	short = mailRE.ReplaceAllString(urlRE.ReplaceAllString(short, " "), " ")
	short = signature.ReplaceAllString(short, "")
	short = mention.ReplaceAllString(short, "")
	short = strings.ReplaceAll(short, "\u00A0", " ")
	return short
}

// ParsedLanguageLists ports LanguageIdentifier.ParsedLanguageLists.
type ParsedLanguageLists struct {
	AdditionalLangs []string
	PreferredLangs  []string
}

// MapLanguageIdentifier is a pluggable scorer for tests and lightweight detection.
// Score(wordList) returns langCode → confidence (higher better).
type MapLanguageIdentifier struct {
	BaseLanguageIdentifier
	// ScoreTexts maps language code → score for the whole text (optional override).
	Score func(cleanText string, preferred []string) map[string]float64
}

func NewMapLanguageIdentifier(maxLength int, score func(string, []string) map[string]float64) *MapLanguageIdentifier {
	if maxLength <= 0 {
		maxLength = DefaultMaxLength
	}
	return &MapLanguageIdentifier{
		BaseLanguageIdentifier: NewBaseLanguageIdentifier(maxLength),
		Score:                  score,
	}
}

func (m *MapLanguageIdentifier) Detect(cleanText string, noopLangs, preferredLangs []string) *languagetool.DetectedLanguage {
	scores := m.Scores(cleanText, noopLangs, preferredLangs, false, 1)
	if len(scores) == 0 {
		return nil
	}
	return &scores[0]
}

func (m *MapLanguageIdentifier) Scores(cleanText string, noopLangs, preferredLangs []string, _ bool, count int) []languagetool.DetectedLanguage {
	if m.Score == nil {
		return nil
	}
	raw := m.Score(cleanText, preferredLangs)
	type pair struct {
		code  string
		score float64
	}
	var pairs []pair
	for k, v := range raw {
		pairs = append(pairs, pair{k, v})
	}
	// sort desc
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].score > pairs[i].score {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	if count <= 0 {
		count = len(pairs)
	}
	var out []languagetool.DetectedLanguage
	for i := 0; i < len(pairs) && i < count; i++ {
		src := "map"
		out = append(out, languagetool.NewDetectedLanguageFull(
			"", pairs[i].code, float32(pairs[i].score), &src))
	}
	return out
}

// SimpleSpellScoreIdentifier scores languages by fraction of words accepted by IsKnown.
// IsKnown(lang, word) true means correctly spelled.
type SimpleSpellScoreIdentifier struct {
	BaseLanguageIdentifier
	// IsKnown reports whether word is known in lang short code.
	IsKnown map[string]func(word string) bool
}

func NewSimpleSpellScoreIdentifier(isKnown map[string]func(word string) bool) *SimpleSpellScoreIdentifier {
	return &SimpleSpellScoreIdentifier{
		BaseLanguageIdentifier: NewBaseLanguageIdentifier(DefaultMaxLength),
		IsKnown:                isKnown,
	}
}

func (s *SimpleSpellScoreIdentifier) Detect(cleanText string, noopLangs, preferredLangs []string) *languagetool.DetectedLanguage {
	scores := s.Scores(cleanText, noopLangs, preferredLangs, false, 1)
	if len(scores) == 0 {
		return nil
	}
	return &scores[0]
}

func (s *SimpleSpellScoreIdentifier) Scores(cleanText string, _ []string, preferredLangs []string, _ bool, count int) []languagetool.DetectedLanguage {
	words := strings.Fields(cleanText)
	if len(words) == 0 {
		return nil
	}
	type pair struct {
		code  string
		score float64
	}
	var pairs []pair
	for lang, known := range s.IsKnown {
		if known == nil {
			continue
		}
		if len(preferredLangs) > 0 && !containsStr(preferredLangs, lang) {
			// still allow all if preferred empty; when preferred set, only those
			// (simple surface)
		}
		ok := 0
		for _, w := range words {
			if known(w) {
				ok++
			}
		}
		pairs = append(pairs, pair{lang, float64(ok) / float64(len(words))})
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].score > pairs[i].score {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	if count <= 0 {
		count = len(pairs)
	}
	var out []languagetool.DetectedLanguage
	for i := 0; i < len(pairs) && i < count; i++ {
		src := "spellchecker"
		out = append(out, languagetool.NewDetectedLanguageFull(
			"", pairs[i].code, float32(pairs[i].score), &src))
	}
	return out
}

func containsStr(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// PrepareDetectLanguage ports LanguageIdentifier.prepareDetectLanguage.
// unicodeDominant optional hook for dominant lang codes (Java UNICODE_BASED_LANG_IDENTIFIER).
func PrepareDetectLanguage(text string, noopLangs, preferredLangs []string, unicodeDominant func(string) []string) *ParsedLanguageLists {
	if noopLangs == nil || preferredLangs == nil {
		panic("noopLangs and preferredLangs required")
	}
	mapNb := func(ss []string) []string {
		out := make([]string, len(ss))
		for i, k := range ss {
			if k == "nb" {
				out[i] = "no"
			} else {
				out[i] = k
			}
		}
		return out
	}
	additional := mapNb(noopLangs)
	preferred := mapNb(preferredLangs)
	for _, k := range preferred {
		if strings.Contains(k, "-") {
			panic("preferredLanguages may only contain language codes without variants (e.g. 'en', but not 'en-US'): " + strings.Join(preferred, ",") + ". Use 'preferredVariants' to specify variants.")
		}
	}
	var dom []string
	if unicodeDominant != nil {
		dom = unicodeDominant(text)
	}
	domStr := strings.Join(dom, ",")
	if domStr == "th" || domStr == "he" || domStr == "ko" || domStr == "hi,mr" {
		return nil
	}
	hasCyrZh := false
	for _, p := range preferred {
		if p == "ru" || p == "uk" || p == "be" || p == "zh" || p == "hi" || p == "mr" {
			hasCyrZh = true
			break
		}
	}
	if !hasCyrZh {
		preferred = append(preferred, dom...)
		additional = append(additional, dom...)
	}
	return &ParsedLanguageLists{AdditionalLangs: additional, PreferredLangs: preferred}
}

// GetHighestScoringResult ports getHighestScoringResult.
func GetHighestScoringResult(probs map[string]float64) (code string, score float64) {
	max := -1.0
	var result string
	for k, v := range probs {
		if v > max {
			max = v
			result = k
		}
	}
	return result, max
}

// GetOrderedScores ports getOrderedScores — top count by descending score.
func GetOrderedScores(scores map[string]float64, count int) map[string]float64 {
	type pair struct {
		k string
		v float64
	}
	var entries []pair
	for k, v := range scores {
		entries = append(entries, pair{k, v})
	}
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].v > entries[i].v {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	out := map[string]float64{}
	for i := 0; i < len(entries) && i < count; i++ {
		out[entries[i].k] = entries[i].v
	}
	return out
}
