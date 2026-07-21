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
	infoRunOnWords        = "fsa.dict.speller.runon-words"
	infoInputConversion   = "fsa.dict.input-conversion"
	infoOutputConversion  = "fsa.dict.output-conversion"
	infoReplacementPairs  = "fsa.dict.speller.replacement-pairs"
	infoEquivalentChars   = "fsa.dict.speller.equivalent-chars"
	infoLocale            = "fsa.dict.speller.locale"
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
	if v, ok := meta[infoRunOnWords]; ok {
		s.SupportRunOnWords = parseInfoBool(v, s.SupportRunOnWords)
	}
	if v, ok := meta[infoIgnoreDiacritics]; ok {
		s.IgnoreDiacritics = parseInfoBool(v, s.IgnoreDiacritics)
	}
	if v, ok := meta[infoEquivalentChars]; ok {
		s.EquivalentChars = ParseEquivalentChars(v)
	}
	if v, ok := meta[infoInputConversion]; ok {
		s.InputConversionPairs = ParseConversionPairs(v)
	}
	if v, ok := meta[infoOutputConversion]; ok {
		s.OutputConversionPairs = ParseConversionPairs(v)
	}
	if v, ok := meta[infoReplacementPairs]; ok {
		rest, short := partitionReplacementPairs(ParseReplacementPairs(v))
		s.ReplacementTheRest = rest
		s.ReplacementShort = short
	}
	if v, ok := meta[infoLocale]; ok {
		s.ConversionLocale = strings.TrimSpace(v)
	}
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

// AttachBinaryDictionary opens/caches FSA at dictPath and wires a per-instance Speller
// (Java MorfologikSpeller holds Speller; Dictionary is cache-shared/immutable).
// Also loads sibling .info flags.
func (s *MorfologikSpeller) AttachBinaryDictionary(dictPath string) bool {
	if s == nil || dictPath == "" {
		return false
	}
	d := openCachedBinaryDict(dictPath)
	if d == nil {
		return false
	}
	s.LoadInfoBesideDict(dictPath)
	// Sync flags LT may have loaded onto Dictionary before building Speller.
	s.syncDictSpellerMeta(d)
	sp := atticmorfo.NewSpeller(d, s.MaxEditDistance)
	sp.SyncFromDict()
	s.binarySpeller = sp
	s.InDictionaryFn = sp.IsInDictionary
	s.BinaryDictPath = dictPath
	s.binaryDict = d
	s.FrequencyIncluded = d.FrequencyIncluded()
	// Binary suggest: Java findReplacementCandidates at MaxEditDistance + replacement-pairs.
	// Rule-level cascade (speller1/2/3) lives in MorfologikSpellerRule.collectSuggestions.
	// Suggest hooks unused by GetWeightedSuggestions when binaryDict is set (direct Speller path).
	// Kept for callers that invoke SuggestFn/WeightedSuggestFn alone — full candidate list (no invent 8-cap).
	s.SuggestFn = func(word string) []string {
		return wordsFromWeighted(s.binaryFindReplacementCandidates(d, word, 0))
	}
	s.WeightedSuggestFn = func(word string) []WeightedSuggestion {
		return s.binaryFindReplacementCandidates(d, word, 0)
	}
	// Binary frequency: Java Speller.getFrequency last payload byte.
	s.GetFrequencyFn = func(word string) int {
		return d.GetFrequency(word)
	}
	return true
}

// syncDictSpellerMeta copies MorfologikSpeller flags onto Dictionary before Speller use.
func (s *MorfologikSpeller) syncDictSpellerMeta(d *atticmorfo.Dictionary) {
	if s == nil || d == nil {
		return
	}
	d.IgnoreDiacritics = s.IgnoreDiacritics
	d.ConvertCase = s.ConvertCase
	d.IgnoreNumbers = s.IgnoreNumbers
	d.IgnorePunctuation = s.IgnorePunctuation
	d.IgnoreCamelCase = s.IgnoreCamelCase
	d.IgnoreAllUppercase = s.IgnoreAllUppercase
	d.SupportRunOnWords = s.SupportRunOnWords
	d.EquivalentChars = s.EquivalentChars
	if s.ConversionLocale != "" {
		d.SetLocale(s.ConversionLocale)
	}
	if len(s.InputConversionPairs) > 0 {
		d.InputConversion = append([][2]string(nil), s.InputConversionPairs...)
	}
	if len(s.OutputConversionPairs) > 0 {
		d.OutputConversion = append([][2]string(nil), s.OutputConversionPairs...)
	}
	if len(s.ReplacementShort) > 0 || (s.ReplacementTheRest != nil && s.ReplacementTheRest.Len() > 0) {
		d.ReplacementShort = make([]atticmorfo.ReplPair, 0, len(s.ReplacementShort))
		for _, p := range s.ReplacementShort {
			d.ReplacementShort = append(d.ReplacementShort, atticmorfo.ReplPair{From: p.From, To: p.To})
		}
		if s.ReplacementTheRest != nil {
			d.ReplacementTheRest = &atticmorfo.OrderedStringListMap{}
			for _, k := range s.ReplacementTheRest.Keys {
				for _, v := range s.ReplacementTheRest.Get(k) {
					d.ReplacementTheRest.Add(k, v)
				}
			}
		}
	}
}

// newSuggestSpeller ports Java MorfologikSpeller.getSuggestions:
//
//	Speller speller = new Speller(dictionary, maxEditDistance);
//
// Fresh Speller each call (HMatrix / containsSeparators not reused across getSuggestions).
// binarySpeller remains for isMisspelled sticky membership (Java this.speller).
func (s *MorfologikSpeller) newSuggestSpeller(d *atticmorfo.Dictionary) *atticmorfo.Speller {
	if s == nil || d == nil {
		return nil
	}
	maxEdit := s.MaxEditDistance
	if maxEdit < 1 {
		maxEdit = 1
	}
	s.syncDictSpellerMeta(d)
	sp := atticmorfo.NewSpeller(d, maxEdit)
	sp.SyncFromDict()
	return sp
}

// binaryFindReplacementCandidates ports Speller.findReplacementCandidates for a single call.
// Always uses a fresh Speller (Java getSuggestions constructs Speller per invocation).
// maxResults <= 0 → no cap (Java returns full candidate list).
func (s *MorfologikSpeller) binaryFindReplacementCandidates(d *atticmorfo.Dictionary, word string, maxResults int) []WeightedSuggestion {
	if s == nil || d == nil || word == "" {
		return nil
	}
	sp := s.newSuggestSpeller(d)
	if sp == nil {
		return nil
	}
	cds := sp.FindReplacementCandidatesFull(word, false)
	if len(cds) == 0 {
		return nil
	}
	out := make([]WeightedSuggestion, 0, len(cds))
	for _, cd := range cds {
		out = append(out, NewWeightedSuggestion(cd.Word, cd.Distance))
		if maxResults > 0 && len(out) >= maxResults {
			break
		}
	}
	return out
}

// binarySuggestWithRunOn ports getSuggestions core: one Speller for findRepl + run-on.
func (s *MorfologikSpeller) binarySuggestWithRunOn(d *atticmorfo.Dictionary, word string) []WeightedSuggestion {
	if s == nil || d == nil || word == "" {
		return nil
	}
	sp := s.newSuggestSpeller(d)
	if sp == nil {
		return nil
	}
	var out []WeightedSuggestion
	// Java: if (word.length() < 50) findReplacementCandidates
	if UTF16Len(word) < morfologikFindReplMaxLen {
		for _, cd := range sp.FindReplacementCandidatesFull(word, false) {
			out = append(out, NewWeightedSuggestion(cd.Word, cd.Distance))
		}
	}
	// same Speller instance (sticky containsSeparators within this getSuggestions)
	for _, cd := range sp.ReplaceRunOnWordCandidates(word) {
		out = append(out, NewWeightedSuggestion(cd.Word, cd.Distance))
	}
	return out
}

func mergeWeightedUnique(a, b []WeightedSuggestion) []WeightedSuggestion {
	seen := map[string]struct{}{}
	var out []WeightedSuggestion
	for _, s := range a {
		if _, ok := seen[s.Word]; ok {
			continue
		}
		seen[s.Word] = struct{}{}
		out = append(out, s)
	}
	for _, s := range b {
		if _, ok := seen[s.Word]; ok {
			continue
		}
		seen[s.Word] = struct{}{}
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
