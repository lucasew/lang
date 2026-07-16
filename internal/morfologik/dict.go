package morfologik

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Dictionary is FSA + metadata (.dict + .info).
type Dictionary struct {
	FSA       *FSA
	Separator byte
	Encoder   string // SUFFIX, PREFIX, INFIX, NONE
	Encoding  string
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
	return &Dictionary{
		FSA:       fsa,
		Separator: sep,
		Encoder:   strings.ToUpper(enc),
		Encoding:  meta["fsa.dict.encoding"],
	}, nil
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

// Lookup returns morphological analyses for word (UTF-8).
func (d *Dictionary) Lookup(word string) ([]WordForm, error) {
	if word == "" {
		return nil, nil
	}
	seq := []byte(word)
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
		stem, err := d.decodeStem(seq, ba[:sepPos])
		if err != nil {
			return nil, err
		}
		tag := ""
		if sepPos < len(ba) {
			tag = string(ba[sepPos+1:])
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
	kind, _, _ := d.FSA.Match([]byte(word), d.FSA.RootNode())
	return kind == ExactMatch
}

func (d *Dictionary) encoderPrefixBytes() int {
	switch d.Encoder {
	case "NONE":
		return 0
	case "SUFFIX", "PREFIX":
		return 1
	case "INFIX", "PREFIX_INFIX", "TRIM_INFIX_AND_SUFFIX":
		return 2 // approximate; INFIX uses more in some versions
	default:
		return 1
	}
}

func (d *Dictionary) decodeStem(source, encoded []byte) (string, error) {
	switch d.Encoder {
	case "NONE":
		return string(encoded), nil
	case "SUFFIX":
		return decodeTrimSuffix(source, encoded), nil
	default:
		// Fall back: treat encoded as raw suffix after first control byte if present
		if len(encoded) == 0 {
			return string(source), nil
		}
		return decodeTrimSuffix(source, encoded), nil
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
