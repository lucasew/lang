package morfologik

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
)

// morfologik DictionaryAttribute property names (fsa.dict.speller.*).
const (
	infoIgnoreNumbers     = "fsa.dict.speller.ignore-numbers"
	infoIgnorePunctuation = "fsa.dict.speller.ignore-punctuation"
	infoIgnoreCamelCase   = "fsa.dict.speller.ignore-camel-case"
	infoIgnoreAllUpper    = "fsa.dict.speller.ignore-all-uppercase"
	infoIgnoreDiacritics  = "fsa.dict.speller.ignore-diacritics"
	infoConvertCase       = "fsa.dict.speller.convert-case"
	infoFrequencyIncluded = "fsa.dict.frequency-included"
)

// binaryDictCache caches OpenDictionary by absolute path (FSA is thread-safe).
var binaryDictCache sync.Map // string -> *atticmorfo.Dictionary

// ApplyInfoProperties ports DictionaryMetadata.fromMap / .info speller flags.
// Unspecified keys leave current values (call after NewMorfologikSpeller defaults).
func (s *MorfologikSpeller) ApplyInfoProperties(meta map[string]string) {
	if s == nil || meta == nil {
		return
	}
	if v, ok := meta[infoIgnoreNumbers]; ok {
		s.IgnoreNumbers = parseInfoBool(v, s.IgnoreNumbers)
	}
	if v, ok := meta[infoIgnorePunctuation]; ok {
		s.IgnorePunctuation = parseInfoBool(v, s.IgnorePunctuation)
	}
	if v, ok := meta[infoIgnoreCamelCase]; ok {
		s.IgnoreCamelCase = parseInfoBool(v, s.IgnoreCamelCase)
	}
	if v, ok := meta[infoIgnoreAllUpper]; ok {
		s.IgnoreAllUppercase = parseInfoBool(v, s.IgnoreAllUppercase)
	}
	if v, ok := meta[infoConvertCase]; ok {
		s.ConvertCase = parseInfoBool(v, s.ConvertCase)
	}
	if v, ok := meta[infoFrequencyIncluded]; ok {
		s.FrequencyIncluded = parseInfoBool(v, s.FrequencyIncluded)
	}
	// ignore-diacritics affects suggestion search, not isMisspelled gates — stored when needed later.
	_ = meta[infoIgnoreDiacritics]
}

// LoadInfoBesideDict reads path.dict's sibling .info and applies speller flags.
// Returns false if .info missing or unreadable (fail closed: keep current flags).
func (s *MorfologikSpeller) LoadInfoBesideDict(dictPath string) bool {
	if s == nil || dictPath == "" {
		return false
	}
	infoPath := strings.TrimSuffix(dictPath, filepath.Ext(dictPath)) + ".info"
	meta, err := readSpellerInfoFile(infoPath)
	if err != nil || len(meta) == 0 {
		return false
	}
	s.ApplyInfoProperties(meta)
	return true
}

// AttachBinaryDictionary opens/caches FSA at dictPath and sets InDictionaryFn to Dictionary.Contains.
// Also loads sibling .info flags. Java MorfologikSpeller(Dictionary) path.
func (s *MorfologikSpeller) AttachBinaryDictionary(dictPath string) bool {
	if s == nil || dictPath == "" {
		return false
	}
	d := openCachedBinaryDict(dictPath)
	if d == nil {
		return false
	}
	s.LoadInfoBesideDict(dictPath)
	s.InDictionaryFn = d.Contains
	s.BinaryDictPath = dictPath
	s.binaryDict = d
	s.FrequencyIncluded = d.FrequencyIncluded()
	// Binary suggest: Java calcSpellerSuggestions cascade speller1→2→3 (edit distance 1,2,3).
	// maxEditDistance on this speller is the primary; cascade uses 1 then 2 then 3 when empty.
	s.SuggestFn = func(word string) []string {
		return binaryCascadeSuggestions(d, word, 8)
	}
	// Binary frequency: Java Speller.getFrequency last payload byte.
	s.GetFrequencyFn = func(word string) int {
		return d.GetFrequency(word)
	}
	return true
}

// binaryCascadeSuggestions ports MorfologikSpellerRule.calcSpellerSuggestions distance cascade:
// edit-1 first; if empty (or only case-differs) and len>=3 use edit-2; if still empty and len>=5 use edit-3.
func binaryCascadeSuggestions(d *atticmorfo.Dictionary, word string, max int) []string {
	if d == nil || word == "" {
		return nil
	}
	// Weighted sort by distance+frequency (Java WeightedSuggestion)
	w1 := d.WeightedEditSuggestions(word, max, 1)
	sugs := weightedWords(w1)
	onlyCase := len(sugs) > 0 && strings.EqualFold(word, sugs[0])
	if len(word) >= 3 && (onlyCase || len(sugs) == 0) {
		w2 := d.WeightedEditSuggestions(word, max, 2)
		sugs = mergeUnique(sugs, weightedWords(w2))
		if len(word) >= 5 && len(sugs) == 0 {
			w3 := d.WeightedEditSuggestions(word, max, 3)
			sugs = mergeUnique(sugs, weightedWords(w3))
		}
	}
	if len(sugs) > max {
		sugs = sugs[:max]
	}
	return sugs
}

func weightedWords(w []struct {
	Word   string
	Weight int
}) []string {
	out := make([]string, 0, len(w))
	for _, x := range w {
		out = append(out, x.Word)
	}
	return out
}

func mergeUnique(a, b []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range a {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	for _, s := range b {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// TryAttachBinaryFromClasspath discovers Java resource classpath (.dict) on disk and attaches.
func (s *MorfologikSpeller) TryAttachBinaryFromClasspath(classpath string) bool {
	if s == nil {
		return false
	}
	path := classpath
	if st, err := os.Stat(classpath); err != nil || !st.Mode().IsRegular() {
		path = DiscoverLanguageDict(classpath)
	}
	if path == "" {
		return false
	}
	return s.AttachBinaryDictionary(path)
}

// LoadInfoFromClasspath discovers dict path and loads .info only (no FSA open).
func (s *MorfologikSpeller) LoadInfoFromClasspath(classpath string) bool {
	if s == nil {
		return false
	}
	path := classpath
	if st, err := os.Stat(classpath); err != nil || !st.Mode().IsRegular() {
		path = DiscoverLanguageDict(classpath)
	}
	if path == "" {
		return false
	}
	return s.LoadInfoBesideDict(path)
}

func openCachedBinaryDict(dictPath string) *atticmorfo.Dictionary {
	if v, ok := binaryDictCache.Load(dictPath); ok {
		if d, ok := v.(*atticmorfo.Dictionary); ok {
			return d
		}
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return nil
	}
	actual, _ := binaryDictCache.LoadOrStore(dictPath, d)
	return actual.(*atticmorfo.Dictionary)
}

func readSpellerInfoFile(path string) (map[string]string, error) {
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

func parseInfoBool(s string, def bool) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return def
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return b
}
