package morfologik

import (
	"strings"
	"unicode"
	"unicode/utf16"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// MorfologikSpeller ports org.languagetool.rules.spelling.morfologik.MorfologikSpeller
// as a pluggable dictionary probe + optional suggestion map (binary .dict deferred).
//
// Metadata flags port morfologik.speller.Speller.isMisspelled gates driven by
// DictionaryMetadata (defaults: ignoreNumbers/punctuation true, convertCase true).
type MorfologikSpeller struct {
	// FileInClassPath is the dictionary resource path (API parity).
	FileInClassPath string
	MaxEditDistance int
	// Words accepted by the speller.
	Words map[string]struct{}
	// Suggestions for misspellings.
	Suggestions map[string][]string
	// Frequencies ports dictionary frequency tags (optional inject; default 1 for known words).
	// Java MorfologikSpeller.getFrequency; used by wrong-split frequency gates.
	Frequencies map[string]int
	// ConversionLocale is fsa.dict.speller.locale (e.g. "en_US") for toLowerCase/toUpperCase.
	ConversionLocale string

	// Speller dictionary metadata (morfologik DictionaryMetadata / .info).
	// Zero value of *bool is not used — flags are concrete with NewMorfologikSpeller defaults.
	IgnoreNumbers      bool // default true — words with digits are not misspelled
	IgnorePunctuation  bool // default true — single non-alphabetic char not misspelled
	IgnoreCamelCase    bool // default true; en_US.info sets false
	IgnoreAllUppercase bool // default true; en_US.info sets false
	ConvertCase        bool // default true — lowercase / title probe

	// InDictionaryFn ports binary FSA membership (Dictionary.Contains / Speller.isInDictionary).
	// When set, HasDictionary is true even if Words is empty.
	InDictionaryFn func(word string) bool
	// SuggestFn ports binary Speller.findReplacements (optional; map Suggestions still preferred).
	SuggestFn func(word string) []string
	// WeightedSuggestFn ports Speller.findReplacementCandidates with distances (optional).
	WeightedSuggestFn func(word string) []WeightedSuggestion
	// GetFrequencyFn ports binary Speller.getFrequency (optional; Frequencies map preferred).
	GetFrequencyFn func(word string) int
	// BinaryDictPath is the absolute .dict path when AttachBinaryDictionary succeeded.
	BinaryDictPath string
	// FrequencyIncluded ports fsa.dict.frequency-included (from .info / binary dict).
	FrequencyIncluded bool
	// SupportRunOnWords ports fsa.dict.speller.runon-words (default true).
	SupportRunOnWords bool
	// IgnoreDiacritics ports fsa.dict.speller.ignore-diacritics (EN true).
	IgnoreDiacritics bool
	// EquivalentChars ports fsa.dict.speller.equivalent-chars (from → alternatives).
	EquivalentChars map[rune][]rune
	// InputConversionPairs ports fsa.dict.input-conversion (ordered).
	InputConversionPairs [][2]string
	// OutputConversionPairs ports fsa.dict.output-conversion (ordered).
	OutputConversionPairs [][2]string
	// ReplacementTheRest ports Speller.replacementsTheRest (multi-char targets, len>=3; LinkedHashMap order).
	ReplacementTheRest *LinkedHashStringListMap
	// ReplacementShort ports anyToOne/anyToTwo pairs (dict target len 1–2) for findRepl maps.
	ReplacementShort []ReplacementPair
	// binaryDict is the attached FSA (typed as any to avoid exporting attic in all call sites).
	// Set only via AttachBinaryDictionary; used for identity/debug.
	binaryDict any
	// binarySpeller is the per-instance morfologik Speller (sticky containsSeparators).
	// Java MorfologikSpeller holds one Speller; Dictionary is shared/immutable.
	binarySpeller *atticmorfo.Speller
}

func NewMorfologikSpeller(fileInClassPath string, maxEditDistance int) *MorfologikSpeller {
	if maxEditDistance < 1 {
		maxEditDistance = 1
	}
	s := &MorfologikSpeller{
		FileInClassPath: fileInClassPath,
		MaxEditDistance: maxEditDistance,
		Words:           map[string]struct{}{},
		Suggestions:     map[string][]string{},
		Frequencies:     map[string]int{},
		// morfologik DictionaryMetadata.DEFAULT_ATTRIBUTES
		IgnoreNumbers:      true,
		IgnorePunctuation:  true,
		IgnoreCamelCase:    true,
		IgnoreAllUppercase: true,
		ConvertCase:        true,
		SupportRunOnWords:  true,
	}
	// Prefer real sibling .info when the Java resource is on disk; else EN hunspell twin.
	if !s.LoadInfoFromClasspath(fileInClassPath) && isEnglishHunspellDict(fileInClassPath) {
		// en_US.info twin when file not found (CI without third_party dicts).
		s.IgnoreCamelCase = false
		s.IgnoreAllUppercase = false
	}
	return s
}

// HasDictionary reports map inject and/or binary FSA membership available.
// Java MorfologikSpeller always has a Dictionary; empty map without binary is fail-closed.
func (s *MorfologikSpeller) HasDictionary() bool {
	if s == nil {
		return false
	}
	if s.InDictionaryFn != nil {
		return true
	}
	return len(s.Words) > 0
}

func isEnglishHunspellDict(path string) bool {
	// Java resource paths like /en/hunspell/en_US.dict
	return strings.Contains(path, "/en/hunspell/") || strings.Contains(path, "en_US.dict") ||
		strings.Contains(path, "en_GB.dict") || strings.Contains(path, "en_CA.dict") ||
		strings.Contains(path, "en_AU.dict") || strings.Contains(path, "en_NZ.dict") ||
		strings.Contains(path, "en_ZA.dict")
}

// AddWord registers an accepted dictionary form.
func (s *MorfologikSpeller) AddWord(word string) {
	if s.Words == nil {
		s.Words = map[string]struct{}{}
	}
	s.Words[word] = struct{}{}
}

// SetFrequency injects a dictionary frequency for wrong-split tests (Java fsa freq tag).
func (s *MorfologikSpeller) SetFrequency(word string, freq int) {
	if s == nil {
		return
	}
	if s.Frequencies == nil {
		s.Frequencies = map[string]int{}
	}
	s.Frequencies[word] = freq
}

// GetFrequency ports MorfologikSpeller.getFrequency (exact then lowercase).
// Java: int freq = speller.getFrequency(word); if (freq == 0 && !word.equals(word.toLowerCase())) ...
// Uses ConversionLocale when set (dict locale); else strings.ToLower like default Locale path.
// Do not invent 1 for known map words (weights: dist*26+26-freq-1 → wordone/51 with freq 0).
func (s *MorfologikSpeller) GetFrequency(word string) int {
	if s == nil || word == "" {
		return 0
	}
	if f, ok := s.lookupFrequency(word); ok {
		// Java returns speller freq even when 0 for exact hit path only if... actually
		// Java always tries lowercase when freq==0, even for known words with tag 0.
		if f > 0 {
			return f
		}
	}
	// Java: if (freq == 0 && !word.equals(word.toLowerCase()))
	low := s.toLower(word)
	if low != word {
		if f, ok := s.lookupFrequency(low); ok {
			return f
		}
	}
	// exact was known with freq 0
	if f, ok := s.lookupFrequency(word); ok {
		return f
	}
	return 0
}

// lookupFrequency returns (freq, true) when word is known to this speller's freq sources.
// true with 0 is valid (rare word / no frequency tags).
func (s *MorfologikSpeller) lookupFrequency(word string) (int, bool) {
	if s.Frequencies != nil {
		if f, ok := s.Frequencies[word]; ok {
			return f, true
		}
	}
	if s.GetFrequencyFn != nil {
		// Binary path: Speller.getFrequency; 0 when unknown OR known with tag 0.
		// Distinguish via inDictionary when FrequencyIncluded.
		f := s.GetFrequencyFn(word)
		if f > 0 {
			return f, true
		}
		if s.inDictionary(word) {
			return f, true // may be 0
		}
		return 0, false
	}
	// Map inject without explicit Frequencies: Java test.dict without frequency → 0
	if s.inDictionary(word) {
		return 0, true
	}
	return 0, false
}

// IsMisspelled ports morfologik.speller.Speller.isMisspelled metadata gates + dictionary probe.
func (s *MorfologikSpeller) IsMisspelled(word string) bool {
	if s == nil || word == "" {
		return false
	}
	// Java SpellingCheckRule.LANGUAGETOOL / LANGUAGETOOLER short-circuit is on MorfologikSpeller (LT).
	if word == "LanguageTool" || word == "LanguageTooler" {
		return false
	}
	// When binary Speller is attached, use sticky Speller.isMisspelled (Java twin).
	if s.binarySpeller != nil {
		if d := s.binarySpeller.Dict; d != nil {
			s.syncDictSpellerMeta(d)
			s.binarySpeller.SyncFromDict()
		}
		return s.binarySpeller.IsMisspelled(word)
	}
	// Dictionary-only path (no Speller yet): cold Dictionary.IsMisspelled.
	if d, ok := s.binaryDict.(*atticmorfo.Dictionary); ok && d != nil {
		s.syncDictSpellerMeta(d)
		return d.IsMisspelled(word)
	}
	// Map-inject / no-FSA path: same gates with Words / InDictionaryFn.
	wordToCheck := s.applyInputConversion(word)
	// Java: isAlphabetic = word.length() != 1 || isAlphabetic(charAt(0))  (UTF-16 length/charAt)
	u := utf16.Encode([]rune(wordToCheck))
	isAlpha := len(u) != 1 || isAlphabeticCodePoint(rune(u[0]))
	if s.IgnorePunctuation && !isAlpha {
		return false
	}
	if s.IgnoreNumbers && containsDigitUTF16(wordToCheck) {
		return false
	}
	if s.IgnoreCamelCase && isCamelCaseWord(wordToCheck) {
		return false
	}
	if s.IgnoreAllUppercase && isAlpha && isAllUppercaseWord(wordToCheck) {
		return false
	}
	if s.inDictionary(wordToCheck) {
		return false
	}
	if s.ConvertCase && !isMixedCaseWord(wordToCheck) {
		low := s.toLower(wordToCheck)
		if s.inDictionary(low) {
			return false
		}
		if isAllUppercaseWord(wordToCheck) {
			if iu := s.initialUppercase(wordToCheck); iu != wordToCheck && s.inDictionary(iu) {
				return false
			}
		}
	}
	return true
}

// toLower ports word.toLowerCase(dictionaryMetadata.getLocale()).
func (s *MorfologikSpeller) toLower(word string) string {
	if word == "" {
		return word
	}
	tag := s.localeTag()
	if tag == language.Und {
		return strings.ToLower(word)
	}
	return cases.Lower(tag).String(word)
}

func (s *MorfologikSpeller) toUpper(word string) string {
	if word == "" {
		return word
	}
	tag := s.localeTag()
	if tag == language.Und {
		return strings.ToUpper(word)
	}
	return cases.Upper(tag).String(word)
}

func (s *MorfologikSpeller) localeTag() language.Tag {
	if s == nil || s.ConversionLocale == "" {
		return language.Und
	}
	tag, err := language.Parse(strings.ReplaceAll(s.ConversionLocale, "_", "-"))
	if err != nil {
		return language.Und
	}
	return tag
}

// initialUppercase ports Speller.initialUppercase with ConversionLocale.
func (s *MorfologikSpeller) initialUppercase(wordToCheck string) string {
	r := []rune(wordToCheck)
	if len(r) == 0 {
		return wordToCheck
	}
	first := s.toUpper(string(r[0]))
	fr := []rune(first)
	if len(fr) == 0 {
		return wordToCheck
	}
	rest := ""
	if len(r) > 1 {
		rest = s.toLower(string(r[1:]))
	}
	return string(fr[0]) + rest
}

func (s *MorfologikSpeller) applyInputConversion(word string) string {
	if s == nil || len(s.InputConversionPairs) == 0 {
		return word
	}
	return ApplyConversionPairs(word, s.InputConversionPairs)
}

func (s *MorfologikSpeller) applyOutputConversion(word string) string {
	if s == nil || len(s.OutputConversionPairs) == 0 {
		return word
	}
	return ApplyConversionPairs(word, s.OutputConversionPairs)
}

func (s *MorfologikSpeller) inDictionary(word string) bool {
	if s == nil || word == "" {
		return false
	}
	if _, ok := s.Words[word]; ok {
		return true
	}
	if s.InDictionaryFn != nil && s.InDictionaryFn(word) {
		return true
	}
	return false
}

// ConvertsCase reports case-folding acceptance (Java MorfologikSpeller.convertsCase).
func (s *MorfologikSpeller) ConvertsCase() bool {
	return s != nil && s.ConvertCase
}

func containsDigitUTF16(word string) bool {
	for _, r := range word {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// isAlphabeticCodePoint ports Speller.isAlphabetic (Unicode letter categories).
func isAlphabeticCodePoint(r rune) bool {
	return unicode.Is(unicode.Lu, r) || unicode.Is(unicode.Ll, r) || unicode.Is(unicode.Lt, r) ||
		unicode.Is(unicode.Lm, r) || unicode.Is(unicode.Lo, r) || unicode.Is(unicode.Nl, r)
}

// isAllUppercaseWord ports Speller.isAllUppercase (true unless a letter is lowercase).
func isAllUppercaseWord(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// isMixedCaseWord ports Speller.isMixedCase.
// Capitalized "Water" is NOT mixed case.
func isMixedCaseWord(s string) bool {
	return !isAllUppercaseWord(s) && isNotCapitalizedWord(s) && isNotAllLowercaseWord(s)
}

func isNotAllLowercaseWord(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return true
		}
	}
	return false
}

// isNotCapitalizedWord ports Speller.isNotCapitalizedWord.
func isNotCapitalizedWord(s string) bool {
	r := []rune(s)
	if len(r) == 0 {
		return true
	}
	if !unicode.IsUpper(r[0]) {
		return true
	}
	for i := 1; i < len(r); i++ {
		if unicode.IsLetter(r[i]) && !unicode.IsLower(r[i]) {
			return true
		}
	}
	return false
}

// isCamelCaseWord ports Speller.isCamelCase.
func isCamelCaseWord(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)
	if isAllUppercaseWord(s) || !isNotCapitalizedWord(s) {
		return false
	}
	if !unicode.IsUpper(r[0]) {
		return false
	}
	if len(r) > 1 && !unicode.IsLower(r[1]) {
		return false
	}
	return isNotAllLowercaseWord(s)
}

// initialUppercaseWord is map-path helper without locale (tests / applyCase).
func initialUppercaseWord(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	r[0] = unicode.ToUpper(r[0])
	for i := 1; i < len(r); i++ {
		r[i] = unicode.ToLower(r[i])
	}
	return string(r)
}

// GetSuggestions is the Java API alias for FindReplacements (string list, no weights).
func (s *MorfologikSpeller) GetSuggestions(word string) []string {
	return s.FindReplacements(word)
}

// FreqRanges ports morfologik Speller.FREQ_RANGES ('Z'-'A'+1 = 26).
const FreqRanges = 26

// stringMatcherMaxMatchLength ports StringMatcher.MAX_MATCH_LENGTH.
const stringMatcherMaxMatchLength = 250

// morfologikFindReplMaxLen ports MorfologikSpeller.getSuggestions: skip findReplacementCandidates
// when word.length() >= 50 ("slow for long words (the limit is arbitrary)").
const morfologikFindReplMaxLen = 50

// GetWeightedSuggestions ports MorfologikSpeller.getSuggestions (WeightedSuggestion list):
// findReplacementCandidates (if len < 50) + replaceRunOnWordCandidates, then case fold.
// Weights match morfologik CandidateData: edit*FREQ_RANGES + FREQ_RANGES - freq - 1.
func (s *MorfologikSpeller) GetWeightedSuggestions(word string) []WeightedSuggestion {
	if s == nil || word == "" {
		return nil
	}
	// Java: word.length() > StringMatcher.MAX_MATCH_LENGTH → empty
	if UTF16Len(word) > stringMatcherMaxMatchLength {
		return nil
	}
	// Inject Suggestions map: preserve injection order (tests stand in for already-ordered Speller hits).
	if sug, ok := s.Suggestions[word]; ok {
		out := make([]WeightedSuggestion, 0, len(sug))
		for i, w := range sug {
			if w == "" || w == word {
				continue
			}
			// ascending weight by position so multi-merge keeps inject order when freqs equal
			out = append(out, NewWeightedSuggestion(w, i))
		}
		// still merge run-ons after inject (Java always calls replaceRunOnWordCandidates)
		out = mergeWeightedUnique(out, s.ReplaceRunOnWordCandidates(word))
		SortByWeight(out)
		return applyCaseToWeighted(s, word, out)
	}

	var out []WeightedSuggestion
	// Java: if (word.length() < 50) { findReplacementCandidates... }
	if UTF16Len(word) < morfologikFindReplMaxLen {
		if s.WeightedSuggestFn != nil {
			out = append(out, s.WeightedSuggestFn(word)...)
		} else if s.SuggestFn != nil {
			for _, w := range s.SuggestFn(word) {
				if w == "" || w == word {
					continue
				}
				out = append(out, NewWeightedSuggestion(w, s.suggestionWeight(word, w)))
			}
		} else if len(s.Words) > 0 {
			// Plain-text / user-dict map path: Java builds runtime FSA + Speller.findRepl.
			// Without CFSA2 builder, score map peers with SpellerED (Oflazer/Damerau twin).
			out = append(out, s.mapWordsSuggestWeighted(word)...)
		}
	}
	// Java MorfologikSpeller.getSuggestions: always add replaceRunOnWordCandidates
	out = mergeWeightedUnique(out, s.ReplaceRunOnWordCandidates(word))
	if len(out) == 0 {
		return nil
	}
	SortByWeight(out)
	return applyCaseToWeighted(s, word, out)
}

// ReplaceRunOnWordCandidates ports morfologik.speller.Speller.replaceRunOnWordCandidates.
// Suggests "prefix suffix" splits when both sides are in the dictionary (distance 1).
func (s *MorfologikSpeller) ReplaceRunOnWordCandidates(original string) []WeightedSuggestion {
	if s == nil || original == "" || !s.SupportRunOnWords {
		return nil
	}
	// Binary Speller: sticky replaceRunOnWordCandidates (Java Speller twin).
	if s.binarySpeller != nil {
		if d := s.binarySpeller.Dict; d != nil {
			s.syncDictSpellerMeta(d)
		}
		cds := s.binarySpeller.ReplaceRunOnWordCandidates(original)
		if len(cds) == 0 {
			return nil
		}
		out := make([]WeightedSuggestion, 0, len(cds))
		for _, c := range cds {
			out = append(out, NewWeightedSuggestion(c.Word, c.Distance))
		}
		return out
	}
	if d, ok := s.binaryDict.(*atticmorfo.Dictionary); ok && d != nil {
		s.syncDictSpellerMeta(d)
		cds := d.ReplaceRunOnWordCandidates(original)
		if len(cds) == 0 {
			return nil
		}
		out := make([]WeightedSuggestion, 0, len(cds))
		for _, c := range cds {
			out = append(out, NewWeightedSuggestion(c.Word, c.Distance))
		}
		return out
	}
	// Map-inject path (tests / no FSA)
	wordToCheck := s.applyInputConversion(original)
	if s.isInDictionaryExact(wordToCheck) {
		return nil
	}
	// Java: for (i = 1; i < wordToCheck.length(); i++) UTF-16
	u := utf16.Encode([]rune(wordToCheck))
	if len(u) < 2 {
		return nil
	}
	var candidates []WeightedSuggestion
	for i := 1; i < len(u); i++ {
		prefix := string(utf16.Decode(u[:i]))
		suffix := string(utf16.Decode(u[i:]))
		suffixOK := s.isInDictionaryExact(suffix)
		if !suffixOK && !isNotCapitalizedWord(suffix) {
			suffixOK = s.isInDictionaryExact(s.toLower(suffix))
		}
		if !suffixOK {
			continue
		}
		if s.isInDictionaryExact(prefix) {
			candidates = append(candidates, s.runOnCandidate(prefix+" "+suffix))
		} else if prefix != "" {
			pr := []rune(prefix)
			if len(pr) > 0 && unicode.IsUpper(pr[0]) && s.isInDictionaryExact(s.toLower(prefix)) {
				candidates = append(candidates, s.runOnCandidate(prefix+" "+suffix))
			}
		}
	}
	return candidates
}

// ReplaceRunOnWords ports Speller.replaceRunOnWords (strings only).
func (s *MorfologikSpeller) ReplaceRunOnWords(original string) []string {
	ws := s.ReplaceRunOnWordCandidates(original)
	if len(ws) == 0 {
		return nil
	}
	out := make([]string, 0, len(ws))
	for _, w := range ws {
		out = append(out, w.Word)
	}
	return out
}

func (s *MorfologikSpeller) runOnCandidate(replacement string) WeightedSuggestion {
	// Java addReplacement → output conversion then CandidateData(replacement, 1)
	replacement = s.applyOutputConversion(replacement)
	return NewWeightedSuggestion(replacement, s.suggestionWeightDist(replacement, 1))
}

// isInDictionaryExact ports Speller.isInDictionary (exact form only, no misspell gates).
func (s *MorfologikSpeller) isInDictionaryExact(word string) bool {
	return s.inDictionary(word)
}

// suggestionWeight computes Java-like candidate weight for sug relative to word.
func (s *MorfologikSpeller) suggestionWeight(word, sug string) int {
	d := editDistance(strings.ToLower(word), strings.ToLower(sug))
	if d < 1 {
		d = 1
	}
	return s.suggestionWeightDist(sug, d)
}

func (s *MorfologikSpeller) suggestionWeightDist(sug string, dist int) int {
	freq := 0
	if s != nil {
		freq = s.GetFrequency(sug)
		if freq < 0 {
			freq = 0
		}
	}
	// weight = dist*FREQ_RANGES + FREQ_RANGES - frequency - 1
	return dist*FreqRanges + FreqRanges - freq - 1
}

// applyCaseToWeighted ports MorfologikSpeller.getSuggestions all-upper / capitalize arms.
func applyCaseToWeighted(s *MorfologikSpeller, word string, suggestions []WeightedSuggestion) []WeightedSuggestion {
	if s == nil || !s.ConvertCase || len(suggestions) == 0 {
		return suggestions
	}
	if isAllUppercaseWord(word) {
		for i := 0; i < len(suggestions); i++ {
			sugg := suggestions[i]
			allUpper := strings.ToUpper(sugg.Word)
			if allUpper == word || isMixedCaseWord(sugg.Word) {
				allUpper = sugg.Word
			}
			// remove duplicates of allUpper
			aux := weightedIndex(suggestions, allUpper)
			if aux > i {
				suggestions = append(suggestions[:aux], suggestions[aux+1:]...)
			}
			if aux > -1 && aux < i {
				suggestions = append(suggestions[:i], suggestions[i+1:]...)
				i--
			} else {
				suggestions[i] = NewWeightedSuggestion(allUpper, sugg.Weight)
			}
		}
		return suggestions
	}
	if startsWithUppercase(word) {
		for i := 0; i < len(suggestions); i++ {
			sugg := suggestions[i]
			upFirst := uppercaseFirstChar(sugg.Word)
			if upFirst == word || isMixedCaseWord(sugg.Word) {
				upFirst = sugg.Word
			}
			aux := weightedIndex(suggestions, upFirst)
			if aux > i {
				suggestions = append(suggestions[:aux], suggestions[aux+1:]...)
			}
			if aux > -1 && aux < i {
				suggestions = append(suggestions[:i], suggestions[i+1:]...)
				i--
			} else {
				suggestions[i] = NewWeightedSuggestion(upFirst, sugg.Weight)
			}
		}
	}
	return suggestions
}

func weightedIndex(suggestions []WeightedSuggestion, word string) int {
	for i, s := range suggestions {
		if s.Word == word {
			return i
		}
	}
	return -1
}

func startsWithUppercase(s string) bool {
	r := []rune(s)
	return len(r) > 0 && unicode.IsUpper(r[0])
}

func uppercaseFirstChar(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// FindReplacements returns suggestions for word.
// Order: inject Suggestions map → binary SuggestFn → map SpellerED peers.
func (s *MorfologikSpeller) FindReplacements(word string) []string {
	// Prefer weighted path so order matches Java getSuggestions / multi merge.
	if ws := s.GetWeightedSuggestions(word); len(ws) > 0 {
		out := make([]string, 0, len(ws))
		for _, w := range ws {
			out = append(out, w.Word)
		}
		return out
	}
	return nil
}

// mapWordsSuggestWeighted scores Words map peers with SpellerED (Oflazer/Damerau).
// Stand-in for Java runtime FSA + Speller.findReplacementCandidates over plain-text lines.
// Caps at 8 results (same as prior map path). Skips when map is huge (>50k) for latency.
func (s *MorfologikSpeller) mapWordsSuggestWeighted(word string) []WeightedSuggestion {
	if s == nil || word == "" || len(s.Words) == 0 {
		return nil
	}
	if len(s.Words) > 50000 {
		return nil
	}
	maxEdit := s.MaxEditDistance
	if maxEdit < 1 {
		maxEdit = 1
	}
	ed := atticmorfo.NewSpellerED(maxEdit)
	ed.IgnoreDiacritics = s.IgnoreDiacritics
	ed.ConvertCase = s.ConvertCase
	ed.EquivalentChars = s.EquivalentChars

	wordU16 := UTF16Len(word)
	var out []WeightedSuggestion
	for w := range s.Words {
		if w == "" || w == word {
			continue
		}
		// length gate (UTF-16): |len(w)-len(word)| > maxEdit cannot be within maxEdit
		dw := UTF16Len(w) - wordU16
		if dw < 0 {
			dw = -dw
		}
		if dw > maxEdit {
			continue
		}
		d := ed.GetEditDistance(word, w)
		if d > 0 && d <= maxEdit {
			out = append(out, NewWeightedSuggestion(w, s.suggestionWeightDist(w, d)))
		}
	}
	SortByWeight(out)
	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

// editDistance is SpellerED distance (used by suggestionWeight for inject path).
func editDistance(a, b string) int {
	maxEdit := len([]rune(a))
	if lb := len([]rune(b)); lb > maxEdit {
		maxEdit = lb
	}
	if maxEdit < 1 {
		maxEdit = 1
	}
	if maxEdit > 10 {
		maxEdit = 10 // bound HMatrix for long inject pairs
	}
	ed := atticmorfo.NewSpellerED(maxEdit)
	return ed.GetEditDistance(a, b)
}
