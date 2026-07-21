package morfologik

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
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

// minWordLengthFindRepl ports Speller.MIN_WORD_LENGTH (skip short variants after first few).
const minWordLengthFindRepl = 4

// binaryFindReplacementCandidates ports morfologik Speller.findReplacementCandidates (2.2.0):
// input conversion → (optional) getAllReplacements theRest variants → dist-0 dict hits
// → one HMatrix + findRepl per variant (anyToOne/anyToTwo short pairs inside FSA walk)
// → sort → output conversion + first-occurrence dedupe.
// Short pairs are NOT surface-rewritten outside findRepl.
func (s *MorfologikSpeller) binaryFindReplacementCandidates(d *atticmorfo.Dictionary, word string, maxResults int) []WeightedSuggestion {
	if s == nil || d == nil || word == "" {
		return nil
	}
	maxEdit := s.MaxEditDistance
	if maxEdit < 1 {
		maxEdit = 1
	}
	word = s.applyInputConversion(word)
	// evenIfWordInDictionary=false
	if len(word) == 0 || len(word) >= atticmorfo.MaxWordLength || d.Contains(word) {
		return nil
	}

	// One Speller for CandidateData weights + findRepl (Java single Speller instance).
	fsaSp := atticmorfo.NewSpellerFSA(d, maxEdit)
	fsaSp.IgnoreDiacritics = s.IgnoreDiacritics
	fsaSp.ConvertCase = s.ConvertCase
	fsaSp.EquivalentChars = s.EquivalentChars
	if len(s.ReplacementShort) > 0 {
		pairs := make([]struct{ From, To string }, len(s.ReplacementShort))
		for i, p := range s.ReplacementShort {
			pairs[i].From, pairs[i].To = p.From, p.To
		}
		fsaSp.LoadReplacementPairs(pairs)
	}

	// Java: if (replacementsTheRest != null && word.length() > 1) getAllReplacements else [word]
	var wordsToCheck []string
	var raw []atticmorfo.CandidateData
	if s.ReplacementTheRest != nil && s.ReplacementTheRest.Len() > 0 && len(word) > 1 {
		for _, wordChecked := range GetAllReplacements(word, s.ReplacementTheRest, 0, 0) {
			if d.Contains(wordChecked) {
				raw = append(raw, fsaSp.MakeCandidateData(wordChecked, 0))
			} else {
				low := strings.ToLower(wordChecked)
				up := strings.ToUpper(wordChecked)
				if d.Contains(low) {
					raw = append(raw, fsaSp.MakeCandidateData(low, 0))
				}
				if d.Contains(up) {
					raw = append(raw, fsaSp.MakeCandidateData(up, 0))
				}
				if len(low) > 1 {
					firstUp := uppercaseFirstChar(low)
					if d.Contains(firstUp) {
						raw = append(raw, fsaSp.MakeCandidateData(firstUp, 0))
					}
				}
			}
			wordsToCheck = append(wordsToCheck, wordChecked)
		}
	} else {
		wordsToCheck = []string{word}
	}

	// Java: hMatrix.reset() once, then findRepl for each variant (shared dirty matrix).
	fsaSp.ResetHMatrix()
	// Java: int i = 1; for (...) { i++; if (i > UPPER_SEARCH_LIMIT) break;
	// if (wordLen < MIN_WORD_LENGTH && i > 2) break; findRepl(...) }
	i := 1
	for _, wordChecked := range wordsToCheck {
		i++
		if i > upperSearchLimit {
			break
		}
		// Java uses UTF-16 char length; BMP EN matches runes for ASCII.
		if len([]rune(wordChecked)) < minWordLengthFindRepl && i > 2 {
			break
		}
		fsaSp.AppendFindRepl(&raw, wordChecked)
	}

	// Collections.sort(candidates) by weighted distance
	sort.SliceStable(raw, func(a, b int) bool {
		return raw[a].Distance < raw[b].Distance
	})
	// output conversion + first occurrence; Java: new CandidateData(replaced, origDistance)
	seen := map[string]struct{}{}
	var candidates []WeightedSuggestion
	for _, cd := range raw {
		replaced := s.applyOutputConversion(cd.Word)
		if replaced == "" || replaced == word {
			continue
		}
		if _, ok := seen[replaced]; ok {
			continue
		}
		seen[replaced] = struct{}{}
		wt := s.suggestionWeightDist(replaced, cd.OrigDistance)
		candidates = append(candidates, NewWeightedSuggestion(replaced, wt))
	}
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
		IgnoreDiacritics:    s.IgnoreDiacritics,
		ConvertCase:         s.ConvertCase,
		EquivalentChars:     s.EquivalentChars,
		SymmetricEquivalent: true, // invent edit-gen path only
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
