package morfologik

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/language"
)

// Dictionary is FSA + metadata (.dict + .info).
type Dictionary struct {
	FSA               *FSA
	Separator         byte
	Encoder           string // SUFFIX, PREFIX, INFIX, NONE
	Encoding          string
	frequencyIncluded bool // fsa.dict.frequency-included
	// Speller metadata from .info (Java DictionaryMetadata / Speller fields).
	IgnoreDiacritics   bool
	ConvertCase        bool
	IgnoreNumbers      bool // default true in many LT dicts
	IgnorePunctuation  bool
	IgnoreCamelCase    bool
	IgnoreAllUppercase bool
	SupportRunOnWords  bool
	// Locale is fsa.dict.speller.locale (e.g. "en_US"); used for toLowerCase/toUpperCase.
	Locale             string
	langTag            language.Tag
	EquivalentChars    map[rune][]rune
	InputConversion    [][2]string // ordered LinkedHashMap pairs
	OutputConversion   [][2]string
	ReplacementShort   []ReplPair // target len 1–2 → anyToOne/anyToTwo
	ReplacementTheRest *OrderedStringListMap
}

// ReplPair is one fsa.dict.speller.replacement-pairs entry (from=misspelled, to=dict form).
type ReplPair struct {
	From string
	To   string
}

// WordForm is one stem+tag analysis.
type WordForm struct {
	Stem string
	Tag  string
}

// OpenDictionary loads path.dict and path.info (or sibling .info).
func OpenDictionary(dictPath string) (*Dictionary, error) {
	infoPath := strings.TrimSuffix(dictPath, filepath.Ext(dictPath)) + ".info"
	meta, err := readInfo(infoPath)
	if err != nil {
		return nil, err
	}
	fsa, err := OpenFSA(dictPath)
	if err != nil {
		return nil, err
	}
	sep := byte('+')
	if s, ok := meta["fsa.dict.separator"]; ok && s != "" {
		// separator is usually a single character like +
		sep = s[0]
	}
	enc := meta["fsa.dict.encoder"]
	if enc == "" {
		enc = "SUFFIX"
	}
	freqInc := strings.EqualFold(meta["fsa.dict.frequency-included"], "true")
	d := &Dictionary{
		FSA:               fsa,
		Separator:         sep,
		Encoder:           strings.ToUpper(enc),
		Encoding:          meta["fsa.dict.encoding"],
		frequencyIncluded: freqInc,
		// Java DictionaryMetadata defaults (common LT .info override)
		ConvertCase:       true,
		IgnoreNumbers:     true,
		SupportRunOnWords: true,
	}
	d.applySpellerInfo(meta)
	return d, nil
}

// applySpellerInfo ports DictionaryMetadata.fromMap speller-related keys.
func (d *Dictionary) applySpellerInfo(meta map[string]string) {
	if d == nil || meta == nil {
		return
	}
	if v, ok := meta["fsa.dict.speller.ignore-diacritics"]; ok {
		d.IgnoreDiacritics = parseInfoBool(v, d.IgnoreDiacritics)
	}
	if v, ok := meta["fsa.dict.speller.convert-case"]; ok {
		d.ConvertCase = parseInfoBool(v, d.ConvertCase)
	}
	if v, ok := meta["fsa.dict.speller.ignore-numbers"]; ok {
		d.IgnoreNumbers = parseInfoBool(v, d.IgnoreNumbers)
	}
	if v, ok := meta["fsa.dict.speller.ignore-punctuation"]; ok {
		d.IgnorePunctuation = parseInfoBool(v, d.IgnorePunctuation)
	}
	if v, ok := meta["fsa.dict.speller.ignore-camel-case"]; ok {
		d.IgnoreCamelCase = parseInfoBool(v, d.IgnoreCamelCase)
	}
	if v, ok := meta["fsa.dict.speller.ignore-all-uppercase"]; ok {
		d.IgnoreAllUppercase = parseInfoBool(v, d.IgnoreAllUppercase)
	}
	if v, ok := meta["fsa.dict.speller.runon-words"]; ok {
		d.SupportRunOnWords = parseInfoBool(v, d.SupportRunOnWords)
	}
	if v, ok := meta["fsa.dict.speller.locale"]; ok {
		d.setLocale(v)
	}
	if v, ok := meta["fsa.dict.speller.equivalent-chars"]; ok {
		d.EquivalentChars = parseEquivalentCharsInfo(v)
	}
	if v, ok := meta["fsa.dict.input-conversion"]; ok {
		d.InputConversion = parseConversionPairsInfo(v)
	}
	if v, ok := meta["fsa.dict.output-conversion"]; ok {
		d.OutputConversion = parseConversionPairsInfo(v)
	}
	if v, ok := meta["fsa.dict.speller.replacement-pairs"]; ok {
		d.ReplacementShort, d.ReplacementTheRest = partitionReplPairsInfo(parseReplacementPairsInfo(v))
	}
}

func parseInfoBool(s string, def bool) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return def
	}
	if s == "true" || s == "1" || s == "yes" {
		return true
	}
	if s == "false" || s == "0" || s == "no" {
		return false
	}
	return def
}

func readInfo(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		m[strings.TrimSpace(line[:eq])] = strings.TrimSpace(line[eq+1:])
	}
	return m, sc.Err()
}

// Lookup returns morphological analyses for word (UTF-8 Go string).
// Query and decoded stem/tag use fsa.dict.encoding from the .info file
// (e.g. koi8-r for russian.dict, UTF-8 for polish.dict).
func (d *Dictionary) Lookup(word string) ([]WordForm, error) {
	if word == "" {
		return nil, nil
	}
	seq, err := d.encodeBytes(word)
	if err != nil || len(seq) == 0 {
		return nil, err
	}
	// separator must not appear in input
	for _, b := range seq {
		if b == d.Separator {
			return nil, nil
		}
	}
	kind, _, node := d.FSA.Match(seq, d.FSA.RootNode())
	if kind != SequenceIsAPrefix {
		return nil, nil
	}
	// Expect separator arc
	arc := d.FSA.getArc(node, d.Separator)
	if arc == 0 || d.FSA.isArcFinal(arc) {
		return nil, nil
	}
	if d.FSA.isArcTerminal(arc) {
		return nil, nil
	}
	start := d.FSA.endNode(arc)
	seqs := d.FSA.CollectFinalSequences(start)
	var out []WordForm
	for _, ba := range seqs {
		// ba = encodedStem [ + tag ]
		sepPos := -1
		// skip prefix bytes of encoder (SUFFIX uses 1)
		prefix := d.encoderPrefixBytes()
		for i := prefix; i < len(ba); i++ {
			if ba[i] == d.Separator {
				sepPos = i
				break
			}
		}
		if sepPos < 0 {
			sepPos = len(ba)
		}
		stemBytes, err := d.decodeStemBytes(seq, ba[:sepPos])
		if err != nil {
			return nil, err
		}
		stem, err := d.decodeString(stemBytes)
		if err != nil {
			return nil, err
		}
		tag := ""
		if sepPos < len(ba) {
			tag, err = d.decodeString(ba[sepPos+1:])
			if err != nil {
				return nil, err
			}
		}
		out = append(out, WordForm{Stem: stem, Tag: tag})
	}
	return out, nil
}

// Contains is a cold membership probe (fresh Speller.containsSeparators=true each call).
// Prefer Speller.IsInDictionary when sticky Java Speller state is required (shared across
// isMisspelled / findRepl on one MorfologikSpeller). Dictionary is cache-shared and must
// not hold mutable Speller fields.
func (d *Dictionary) Contains(word string) bool {
	return d.IsInDictionary(word)
}

// IsInDictionary is a cold Speller.isInDictionary (per-call local containsSeparators).
// For sticky Java field mutation, use NewSpeller(d, n).IsInDictionary.
func (d *Dictionary) IsInDictionary(word string) bool {
	if d == nil || d.FSA == nil || word == "" {
		return false
	}
	// Cold probe: same control flow as SpellerFSA.IsInDictionary with containsSeparators
	// starting true and discarded after the call (no shared Dictionary mutation).
	sp := &SpellerFSA{Dict: d, containsSeparators: true}
	return sp.IsInDictionary(word)
}

// SetLocale ports DictionaryAttribute.LOCALE (e.g. "en_US" → language tag).
func (d *Dictionary) SetLocale(raw string) {
	d.setLocale(raw)
}

func (d *Dictionary) setLocale(raw string) {
	if d == nil {
		return
	}
	raw = strings.TrimSpace(raw)
	d.Locale = raw
	if raw == "" {
		d.langTag = language.Und
		return
	}
	// Java Locale: language_COUNTRY; BCP-47 uses language-REGION
	tag, err := language.Parse(strings.ReplaceAll(raw, "_", "-"))
	if err != nil {
		d.langTag = language.Und
		return
	}
	d.langTag = tag
}

// ToLower ports word.toLowerCase(dictionaryMetadata.getLocale()).
func (d *Dictionary) ToLower(s string) string {
	if s == "" {
		return s
	}
	if d == nil || d.langTag == language.Und {
		return strings.ToLower(s)
	}
	return cases.Lower(d.langTag).String(s)
}

// ToUpper ports word.toUpperCase(dictionaryMetadata.getLocale()).
func (d *Dictionary) ToUpper(s string) string {
	if s == "" {
		return s
	}
	if d == nil || d.langTag == language.Und {
		return strings.ToUpper(s)
	}
	return cases.Upper(d.langTag).String(s)
}

// encodeBytes converts a UTF-8 Go string to the dictionary's byte encoding.
func (d *Dictionary) encodeBytes(s string) ([]byte, error) {
	enc := d.charset()
	if enc == nil {
		return []byte(s), nil
	}
	return enc.NewEncoder().Bytes([]byte(s))
}

// decodeString converts dictionary bytes to a UTF-8 Go string.
func (d *Dictionary) decodeString(b []byte) (string, error) {
	enc := d.charset()
	if enc == nil {
		return string(b), nil
	}
	out, err := enc.NewDecoder().Bytes(b)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// charset returns the encoding for fsa.dict.encoding (nil = raw UTF-8 bytes).
func (d *Dictionary) charset() encoding.Encoding {
	if d == nil {
		return nil
	}
	name := strings.ToLower(strings.TrimSpace(d.Encoding))
	switch name {
	case "", "utf-8", "utf8":
		return nil
	case "koi8-r", "koi8r":
		return charmap.KOI8R
	case "iso-8859-1", "iso8859-1", "latin1", "latin-1":
		return charmap.ISO8859_1
	case "iso-8859-2", "iso8859-2", "latin2":
		return charmap.ISO8859_2
	case "windows-1250", "cp1250":
		return charmap.Windows1250
	case "windows-1251", "cp1251":
		return charmap.Windows1251
	case "windows-1252", "cp1252":
		return charmap.Windows1252
	case "iso-8859-15", "iso8859-15", "latin9", "latin-9":
		return charmap.ISO8859_15
	case "iso-8859-7", "iso8859-7", "greek":
		// Greek hunspell el_GR.dict (Java Charset ISO-8859-7).
		return charmap.ISO8859_7
	case "iso-8859-9", "iso8859-9", "latin5":
		return charmap.ISO8859_9
	default:
		// Unknown: treat as UTF-8 bytes (Java Charset.forName would throw).
		_ = unicode.UTF8
		return nil
	}
}

func (d *Dictionary) encoderPrefixBytes() int {
	// Matches ISequenceEncoder.prefixBytes() in morfologik-stemming:
	// SUFFIX/NONE: 1/0 control bytes; PREFIX (TrimPrefixAndSuffix): 2; INFIX: 3.
	switch d.Encoder {
	case "NONE":
		return 0
	case "SUFFIX":
		return 1
	case "PREFIX":
		return 2
	case "INFIX", "PREFIX_INFIX", "TRIM_INFIX_AND_SUFFIX":
		return 3
	default:
		return 1
	}
}

// decodeStemBytes returns stem bytes in the dictionary encoding (before charset decode).
func (d *Dictionary) decodeStemBytes(source, encoded []byte) ([]byte, error) {
	switch d.Encoder {
	case "NONE":
		return encoded, nil
	case "SUFFIX":
		return []byte(decodeTrimSuffix(source, encoded)), nil
	case "PREFIX":
		return []byte(decodeTrimPrefixAndSuffix(source, encoded)), nil
	case "INFIX", "PREFIX_INFIX", "TRIM_INFIX_AND_SUFFIX":
		return []byte(decodeTrimInfixAndSuffix(source, encoded)), nil
	default:
		if len(encoded) == 0 {
			return source, nil
		}
		return []byte(decodeTrimSuffix(source, encoded)), nil
	}
}

// removeEverything ports morfologik REMOVE_EVERYTHING (255) in sequence encoders.
const removeEverything = 255

// decodeTrimSuffix implements TrimSuffixEncoder.decode.
// encoded: {K}{suffix} where K-'A' bytes trimmed from source end, then suffix appended.
func decodeTrimSuffix(source, encoded []byte) string {
	// Java: assert encoded.remaining() >= 1
	if len(encoded) < 1 {
		return string(source)
	}
	truncate := int(encoded[0]-'A') & 0xFF
	if truncate == removeEverything {
		truncate = len(source)
	}
	// Java does not clamp; defensive only when corrupt data would panic
	if truncate > len(source) {
		truncate = len(source)
	}
	len1 := len(source) - truncate
	stem := make([]byte, 0, len1+len(encoded)-1)
	stem = append(stem, source[:len1]...)
	stem = append(stem, encoded[1:]...)
	return string(stem)
}

// decodeTrimPrefixAndSuffix implements TrimPrefixAndSuffixEncoder.decode (PREFIX).
// encoded: {P}{K}{suffix} — drop P bytes from start and K from end of source, then append suffix.
func decodeTrimPrefixAndSuffix(source, encoded []byte) string {
	// Java: assert encoded.remaining() >= 2
	if len(encoded) < 2 {
		return string(source)
	}
	truncatePrefix := int(encoded[0]-'A') & 0xFF
	truncateSuffix := int(encoded[1]-'A') & 0xFF
	if truncatePrefix == removeEverything || truncateSuffix == removeEverything {
		truncatePrefix = len(source)
		truncateSuffix = 0
	}
	len1 := len(source) - (truncateSuffix + truncatePrefix)
	if len1 < 0 {
		len1 = 0
	}
	if truncatePrefix > len(source) {
		truncatePrefix = len(source)
		len1 = 0
	}
	stem := make([]byte, 0, len1+len(encoded)-2)
	// Java: put(source, truncatePrefixBytes, len1)
	end := truncatePrefix + len1
	if end > len(source) {
		end = len(source)
	}
	if truncatePrefix < end {
		stem = append(stem, source[truncatePrefix:end]...)
	}
	stem = append(stem, encoded[2:]...)
	return string(stem)
}

// decodeTrimInfixAndSuffix implements TrimInfixAndSuffixEncoder.decode (INFIX).
// encoded: {X}{L}{K}{suffix} — remove L bytes at index X from source, trim K suffix, append suffix.
// See morfologik TrimInfixAndSuffixEncoder (prefixBytes=3).
func decodeTrimInfixAndSuffix(source, encoded []byte) string {
	// Java: assert encoded.remaining() >= 3
	if len(encoded) < 3 {
		// Corrupt/short payload — fail closed like incomplete encode (no invent PREFIX downgrade).
		return ""
	}
	infixIndex := int(encoded[0]-'A') & 0xFF
	infixLength := int(encoded[1]-'A') & 0xFF
	truncateSuffixBytes := int(encoded[2]-'A') & 0xFF

	if infixLength == removeEverything || truncateSuffixBytes == removeEverything {
		// Java: infixIndex=0; infixLength=source.remaining(); truncateSuffixBytes=0
		infixIndex = 0
		infixLength = len(source)
		truncateSuffixBytes = 0
	}

	// len1 = source.remaining() - (infixIndex + infixLength + truncateSuffixBytes)
	len1 := len(source) - (infixIndex + infixLength + truncateSuffixBytes)
	if len1 < 0 {
		len1 = 0
	}
	len2 := len(encoded) - 3
	stem := make([]byte, 0, infixIndex+len1+len2)

	// put(source, 0, infixIndex)
	if infixIndex > len(source) {
		infixIndex = len(source)
	}
	stem = append(stem, source[:infixIndex]...)

	// put(source, infixIndex + infixLength, len1)
	midStart := infixIndex + infixLength
	if midStart > len(source) {
		midStart = len(source)
	}
	midEnd := midStart + len1
	if midEnd > len(source) {
		midEnd = len(source)
	}
	if midStart < midEnd {
		stem = append(stem, source[midStart:midEnd]...)
	}

	// put(encoded, 3, len2)
	stem = append(stem, encoded[3:]...)
	return string(stem)
}

// MustOpen is like OpenDictionary but panics — for tests only via helper.
func MustOpen(path string) *Dictionary {
	d, err := OpenDictionary(path)
	if err != nil {
		panic(err)
	}
	return d
}

// ResolveDict prefers third_party overlay then LT module path.
func ResolveDict(dataRoot, rel string) string {
	// rel e.g. "en/english.dict" under resource/
	candidates := []string{
		filepath.Join(dataRoot, "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource", rel),
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", rel),
	}
	// also relative to dataRoot parent repo
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return candidates[0]
}

// EnsureDictPath returns path or error if missing.
func EnsureDictPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("dictionary not found: %s (run scripts/fetch-english-dicts.sh)", path)
	}
	return nil
}

// FIRST_RANGE_CODE ports morfologik Speller.FIRST_RANGE_CODE ('A').
// Frequency byte is last payload byte after separator: freq = byte - 'A' (0..25).
const firstRangeCode = 'A'

// FrequencyIncluded is set from fsa.dict.frequency-included in .info.
func (d *Dictionary) SetFrequencyIncluded(v bool) {
	if d != nil {
		d.frequencyIncluded = v
	}
}

// FrequencyIncluded reports whether the dict encodes frequency tags.
func (d *Dictionary) FrequencyIncluded() bool {
	return d != nil && d.frequencyIncluded
}

// GetFrequency ports morfologik Speller.getFrequency for speller dictionaries.
// Returns 0 when frequency not included or word unknown.
func (d *Dictionary) GetFrequency(word string) int {
	if d == nil || word == "" || !d.frequencyIncluded {
		return 0
	}
	seq, err := d.encodeBytes(word)
	if err != nil || len(seq) == 0 {
		return 0
	}
	for _, b := range seq {
		if b == d.Separator {
			return 0
		}
	}
	kind, _, node := d.FSA.Match(seq, d.FSA.RootNode())
	if kind != SequenceIsAPrefix {
		return 0
	}
	// Java: arc = fsa.getArc(match.node, separator); arc != 0 && !isArcFinal(arc)
	arc := d.FSA.getArc(node, d.Separator)
	if arc == 0 || d.FSA.isArcFinal(arc) {
		return 0
	}
	if d.FSA.isArcTerminal(arc) {
		return 0
	}
	start := d.FSA.endNode(arc)
	seqs := d.FSA.CollectFinalSequences(start)
	if len(seqs) == 0 {
		return 0
	}
	ba := seqs[0]
	if len(ba) == 0 {
		return 0
	}
	// last byte contains the frequency after a separator (Java)
	return int(ba[len(ba)-1]) - firstRangeCode
}

// NewDictionaryFromWords builds a speller Dictionary from a word list (Java
// MorfologikMultiSpeller runtime FSABuilder.build(lines) + Dictionary.read metadata).
// info may be nil (defaults); typically load sibling .info flags for gates.
func NewDictionaryFromWords(words []string, info map[string]string) *Dictionary {
	fsa := BuildFSAFromWords(words)
	if fsa == nil {
		return nil
	}
	d := &Dictionary{
		FSA:               fsa,
		Separator:         '+',
		Encoder:           "NONE",
		Encoding:          "utf-8",
		frequencyIncluded: false,
		ConvertCase:       true,
		IgnoreNumbers:     true,
		SupportRunOnWords: true,
	}
	if info != nil {
		if s, ok := info["fsa.dict.separator"]; ok && s != "" {
			d.Separator = s[0]
		}
		d.applySpellerInfo(info)
	}
	return d
}
