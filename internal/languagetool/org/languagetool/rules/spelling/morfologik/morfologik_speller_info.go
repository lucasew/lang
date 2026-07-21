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
	// Binary suggest: Java findReplacementCandidates at MaxEditDistance + replacement-pairs.
	// Rule-level cascade (speller1/2/3) lives in MorfologikSpellerRule.collectSuggestions.
	s.SuggestFn = func(word string) []string {
		return wordsFromWeighted(s.binaryFindReplacementCandidates(d, word, 8))
	}
	s.WeightedSuggestFn = func(word string) []WeightedSuggestion {
		return s.binaryFindReplacementCandidates(d, word, 8)
	}
	// Binary frequency: Java Speller.getFrequency last payload byte.
	s.GetFrequencyFn = func(word string) int {
		return d.GetFrequency(word)
	}
	return true
}

// binaryFindReplacementCandidates ports Speller.findReplacementCandidates for one maxEdit:
// input conversion → getAllReplacements variants → distance-0 dict hits + edit search per variant
// → output conversion → sort/dedupe.
func (s *MorfologikSpeller) binaryFindReplacementCandidates(d *atticmorfo.Dictionary, word string, maxResults int) []WeightedSuggestion {
	if s == nil || d == nil || word == "" {
		return nil
	}
	maxEdit := s.MaxEditDistance
	if maxEdit < 1 {
		maxEdit = 1
	}
	if maxEdit > 3 {
		maxEdit = 3
	}
	word = s.applyInputConversion(word)
	// evenIfWordInDictionary=false: empty when already known (caller usually only for misspellings)
	if d.Contains(word) {
		return nil
	}
	// Multi-char theRest variants seed edit search (Java getAllReplacements → findRepl).
	variants := GetAllReplacements(word, s.ReplacementTheRest, 0, 0)
	if len(variants) == 0 {
		variants = []string{word}
	}
	// Short anyToOne/Two: apply only on the original misspelling. Pure rewrite → distance 0
	// if in dict. Do NOT stack short rewrites onto theRest variants or feed them into edit
	// search — that invents multi-step paths outside a single HMatrix budget.
	shortHits := ShortReplacementVariants(word, s.ReplacementShort)
	var candidates []WeightedSuggestion
	seen := map[string]struct{}{}
	addCand := func(w string, dist int) {
		w = s.applyOutputConversion(w)
		if w == "" || w == word {
			return
		}
		if _, ok := seen[w]; ok {
			return
		}
		seen[w] = struct{}{}
		candidates = append(candidates, NewWeightedSuggestion(w, s.suggestionWeightDist(w, dist)))
	}
	for _, h := range shortHits {
		if d.Contains(h) {
			addCand(h, 0)
		} else {
			low := strings.ToLower(h)
			if d.Contains(low) {
				addCand(low, 0)
			} else if firstUp := uppercaseFirstChar(low); firstUp != low && d.Contains(firstUp) {
				addCand(firstUp, 0)
			}
		}
	}
	// Java: for each replacement variant — if in dict, distance 0; always queue for edit search
	i := 0
	for _, wordChecked := range variants {
		i++
		if i > upperSearchLimit {
			break
		}
		if d.Contains(wordChecked) {
			addCand(wordChecked, 0)
		} else {
			low := strings.ToLower(wordChecked)
			up := strings.ToUpper(wordChecked)
			if d.Contains(low) {
				addCand(low, 0)
			}
			if d.Contains(up) {
				addCand(up, 0)
			}
			if len(low) > 1 {
				firstUp := uppercaseFirstChar(low)
				if d.Contains(firstUp) {
					addCand(firstUp, 0)
				}
			}
		}
		// edit-distance search: Java findRepl FSA walk (HMatrix/Oflazer) via SpellerFSA.
		// skip very short after first (Java MIN_WORD_LENGTH=4 && i>2)
		if len([]rune(wordChecked)) < 4 && i > 2 {
			continue
		}
		fsaSp := atticmorfo.NewSpellerFSA(d, maxEdit)
		fsaSp.IgnoreDiacritics = s.IgnoreDiacritics
		fsaSp.EquivalentChars = s.EquivalentChars
		for _, e := range fsaSp.FindReplacementCandidates(wordChecked) {
			w := s.applyOutputConversion(e.Word)
			if w == "" || w == word {
				continue
			}
			if _, ok := seen[w]; ok {
				continue
			}
			seen[w] = struct{}{}
			wt := e.Distance
			if w != e.Word {
				// output conversion changed surface — recompute weight for new form
				wt = s.suggestionWeightDist(w, e.OrigDistance)
				if e.OrigDistance < 1 {
					wt = s.suggestionWeightDist(w, 1)
				}
			}
			candidates = append(candidates, NewWeightedSuggestion(w, wt))
		}
	}
	SortByWeight(candidates)
	if maxResults > 0 && len(candidates) > maxResults {
		candidates = candidates[:maxResults]
	}
	return candidates
}

// suggestOpts builds attic morfologik.SuggestOpts from .info flags.
func (s *MorfologikSpeller) suggestOpts() atticmorfo.SuggestOpts {
	if s == nil {
		return atticmorfo.SuggestOpts{}
	}
	return atticmorfo.SuggestOpts{
		IgnoreDiacritics: s.IgnoreDiacritics,
		EquivalentChars:  s.EquivalentChars,
	}
}

// binaryCascadeWeighted ports calcSpellerSuggestions distance cascade at the binary layer
// (used when a single Speller stands in for speller1+2+3 without Multis).
func binaryCascadeWeighted(d *atticmorfo.Dictionary, word string, max int) []WeightedSuggestion {
	if d == nil || word == "" {
		return nil
	}
	w1 := d.WeightedEditSuggestions(word, max, 1)
	sugs := toWeighted(w1)
	onlyCase := len(sugs) > 0 && strings.EqualFold(word, sugs[0].Word)
	if len(word) >= 3 && (onlyCase || len(sugs) == 0) {
		w2 := d.WeightedEditSuggestions(word, max, 2)
		sugs = mergeWeightedUnique(sugs, toWeighted(w2))
		if len(word) >= 5 && (len(sugs) == 0 || onlyCase) {
			// Java: speller3 only when fullResults || defaultSuggestions.isEmpty() after speller2.
			// onlyCase may leave non-empty list — then speller3 is skipped unless empty.
			if len(sugs) == 0 {
				w3 := d.WeightedEditSuggestions(word, max, 3)
				sugs = mergeWeightedUnique(sugs, toWeighted(w3))
			}
		}
	}
	if len(sugs) > max {
		sugs = sugs[:max]
	}
	return sugs
}

func toWeighted(w []struct {
	Word   string
	Weight int
}) []WeightedSuggestion {
	out := make([]WeightedSuggestion, 0, len(w))
	for _, x := range w {
		out = append(out, NewWeightedSuggestion(x.Word, x.Weight))
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
