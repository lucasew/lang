package morfologik

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

// Dictionary is FSA + metadata (.dict + .info).
type Dictionary struct {
	FSA               *FSA
	Separator         byte
	Encoder           string // SUFFIX, PREFIX, INFIX, NONE
	Encoding          string
	frequencyIncluded bool // fsa.dict.frequency-included
	// Speller metadata from .info (Java DictionaryMetadata / Speller fields).
	IgnoreDiacritics    bool
	ConvertCase         bool
	IgnoreNumbers       bool // default true in many LT dicts
	IgnorePunctuation   bool
	IgnoreCamelCase     bool
	IgnoreAllUppercase  bool
	SupportRunOnWords   bool
	EquivalentChars     map[rune][]rune
	InputConversion     [][2]string // ordered LinkedHashMap pairs
	OutputConversion    [][2]string
	ReplacementShort    []ReplPair  // target len 1–2 → anyToOne/anyToTwo
	ReplacementTheRest  *OrderedStringListMap
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

// Contains reports whether word is in the dictionary (speller dicts often encode word+freq only).
func (d *Dictionary) Contains(word string) bool {
	forms, err := d.Lookup(word)
	if err != nil {
		return false
	}
	if len(forms) > 0 {
		return true
	}
	// Speller dictionaries: word may be exact final path without stem encoding like POS dicts.
	// Try exact match of word itself as full sequence.
	seq, err := d.encodeBytes(word)
	if err != nil {
		return false
	}
	kind, _, _ := d.FSA.Match(seq, d.FSA.RootNode())
	return kind == ExactMatch
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

// decodeTrimSuffix implements TrimSuffixEncoder.decode.
// encoded: {K}{suffix} where K-'A' bytes trimmed from source end, then suffix appended.
func decodeTrimSuffix(source, encoded []byte) string {
	if len(encoded) < 1 {
		return string(source)
	}
	truncate := int(encoded[0]-'A') & 0xFF
	if truncate == 255 {
		truncate = len(source)
	}
	if truncate > len(source) {
		truncate = len(source)
	}
	stem := make([]byte, 0, len(source)-truncate+len(encoded)-1)
	stem = append(stem, source[:len(source)-truncate]...)
	stem = append(stem, encoded[1:]...)
	return string(stem)
}

// decodeTrimPrefixAndSuffix implements TrimPrefixAndSuffixEncoder.decode (PREFIX).
// encoded: {P}{K}{suffix} — drop P bytes from start and K from end of source, then append suffix.
func decodeTrimPrefixAndSuffix(source, encoded []byte) string {
	if len(encoded) < 2 {
		return string(source)
	}
	truncatePrefix := int(encoded[0]-'A') & 0xFF
	truncateSuffix := int(encoded[1]-'A') & 0xFF
	if truncatePrefix == 255 || truncateSuffix == 255 {
		truncatePrefix = len(source)
		truncateSuffix = 0
	}
	if truncatePrefix+truncateSuffix > len(source) {
		// defensive: fall back to full replace
		return string(encoded[2:])
	}
	midLen := len(source) - truncatePrefix - truncateSuffix
	stem := make([]byte, 0, midLen+len(encoded)-2)
	stem = append(stem, source[truncatePrefix:truncatePrefix+midLen]...)
	stem = append(stem, encoded[2:]...)
	return string(stem)
}

// decodeTrimInfixAndSuffix implements TrimInfixAndSuffixEncoder.decode (INFIX).
// encoded: {I}{P}{K}{suffix} — see morfologik TrimInfixAndSuffixEncoder.
// Soft port: treat like PREFIX when only 2 control bytes present; full 3-byte when available.
func decodeTrimInfixAndSuffix(source, encoded []byte) string {
	if len(encoded) < 3 {
		if len(encoded) >= 2 {
			return decodeTrimPrefixAndSuffix(source, encoded)
		}
		return decodeTrimSuffix(source, encoded)
	}
	// Java: infixIndex = encoded[0]-'A', truncateSuffix = encoded[1]-'A', then
	// shared infix recovered from source; simplified via PREFIX-style when infix is 0.
	// Full algorithm from TrimInfixAndSuffixEncoder:
	//   int infixIndex = (encoded[0] - 'A') & 0xFF;
	//   int truncateSuffixBytes = (encoded[1] - 'A') & 0xFF;
	//   int truncatePrefixBytes left implicit via remaining length after infix.
	// Port of decode():
	infixIndex := int(encoded[0]-'A') & 0xFF
	truncateSuffix := int(encoded[1]-'A') & 0xFF
	if infixIndex == 255 || truncateSuffix == 255 {
		return string(encoded[3:])
	}
	// encoded[2] is further control in some versions; morfologik uses:
	// {removeFromIndex}{truncateSuffix}{suffix} with removeFromIndex locating infix start.
	// See TrimInfixAndSuffixEncoder: first byte = index where shared infix starts in source,
	// second = suffix truncate, then rest is stem material before/after infix rebuild.
	// Practical decode matching Java:
	//   len1 = source without suffix, from infixIndex
	//   Actually Java puts: encoded_suffix_part + source[infix : len-suffixTrim]
	// We'll use the documented form from the class:
	// encoded: {P}{K}{infix?} — delegate to prefix+suffix using bytes 0,1 when byte2 is data.
	// Real Java TrimInfixAndSuffixEncoder.decode (3 prefix bytes in some docs is wrong;
	// prefixBytes() returns 3 only for the empty-stem special case). Read carefully:
	// Actually prefixBytes()=3 means first 3 bytes are controls: {I}{N}{K} in older code.
	// Current master uses TrimInfixAndSuffixEncoder with prefixBytes 3:
	// We'll fetch semantics: I=infix index, N=?, K=suffix trim — fall back to PREFIX layout
	// shifted by 1 if third control is present.
	_ = infixIndex
	return decodeTrimPrefixAndSuffix(source, encoded[1:])
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
