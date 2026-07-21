package morfologik

import (
	"strings"
	"unicode"
	"unicode/utf16"
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
	// ConversionLocale lowercases via strings.ToLower when set.
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
	// GetFrequencyFn ports binary Speller.getFrequency (optional; Frequencies map preferred).
	GetFrequencyFn func(word string) int
	// BinaryDictPath is the absolute .dict path when AttachBinaryDictionary succeeded.
	BinaryDictPath string
	// FrequencyIncluded ports fsa.dict.frequency-included (from .info / binary dict).
	FrequencyIncluded bool
	// binaryDict is the attached FSA (typed as any to avoid exporting attic in all call sites).
	// Set only via AttachBinaryDictionary; used for identity/debug.
	binaryDict any
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
// Order: inject Frequencies map → binary GetFrequencyFn → known map/binary membership → 1 → 0.
func (s *MorfologikSpeller) GetFrequency(word string) int {
	if s == nil || word == "" {
		return 0
	}
	if s.Frequencies != nil {
		if f, ok := s.Frequencies[word]; ok {
			return f
		}
	}
	// Java MorfologikSpeller.getFrequency: speller.getFrequency then lowercase
	if s.GetFrequencyFn != nil {
		if f := s.GetFrequencyFn(word); f > 0 {
			return f
		}
		low := strings.ToLower(word)
		if low != word {
			if f := s.GetFrequencyFn(low); f > 0 {
				return f
			}
		}
	}
	if s.inDictionary(word) {
		return 1 // map inject without explicit freq (or binary with freq 0)
	}
	low := strings.ToLower(word)
	if low != word {
		if s.Frequencies != nil {
			if f, ok := s.Frequencies[low]; ok {
				return f
			}
		}
		if s.inDictionary(low) {
			return 1
		}
	}
	return 0
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
	// Java: isAlphabetic = word.length() != 1 || isAlphabetic(charAt(0))  (UTF-16 length/charAt)
	u := utf16.Encode([]rune(word))
	isAlpha := len(u) != 1 || isAlphabeticCodePoint(rune(u[0]))
	// (!ignorePunctuation || isAlphabetic) — single non-letter ignored when ignorePunctuation
	if s.IgnorePunctuation && !isAlpha {
		return false
	}
	// (!ignoreNumbers || containsNoDigit)
	if s.IgnoreNumbers && containsDigitUTF16(word) {
		return false
	}
	if s.IgnoreCamelCase && isCamelCaseWord(word) {
		return false
	}
	if s.IgnoreAllUppercase && isAlpha && isAllUppercaseWord(word) {
		return false
	}
	if s.inDictionary(word) {
		return false
	}
	// convertCase: accept lowercase / initial-upper forms of non-mixed-case words
	if s.ConvertCase && !isMixedCaseWord(word) {
		low := strings.ToLower(word)
		if s.inDictionary(low) {
			return false
		}
		if isAllUppercaseWord(word) {
			// initialUppercase: first upper, rest lower
			if iu := initialUppercaseWord(word); iu != word && s.inDictionary(iu) {
				return false
			}
		}
	}
	return true
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

func isAllUppercaseWord(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

// isMixedCaseWord ports Speller.isMixedCase:
// !isAllUppercase && isNotCapitalizedWord && isNotAllLowercase
// Capitalized "Water" is NOT mixed case.
func isMixedCaseWord(s string) bool {
	if s == "" {
		return false
	}
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

// isNotCapitalizedWord ports Speller.isNotCapitalizedWord (UTF-16 charAt approx via runes for BMP).
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

// isCamelCaseWord ports Speller.isCamelCase (simplified: internal uppercase after lowercase).
func isCamelCaseWord(s string) bool {
	// morfologik: more than one capital, not all upper, first may be upper
	runes := []rune(s)
	if len(runes) < 2 {
		return false
	}
	if isAllUppercaseWord(s) {
		return false
	}
	// at least one upper not at start after a lower
	seenLower := false
	for i, r := range runes {
		if unicode.IsLower(r) {
			seenLower = true
			continue
		}
		if unicode.IsUpper(r) && seenLower && i > 0 {
			return true
		}
	}
	return false
}

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

// GetSuggestions is the Java API alias for FindReplacements.
func (s *MorfologikSpeller) GetSuggestions(word string) []string {
	return s.FindReplacements(word)
}

// FindReplacements returns suggestions for word.
// Order: inject Suggestions map → binary SuggestFn → small-map edit-distance peers.
func (s *MorfologikSpeller) FindReplacements(word string) []string {
	if s == nil {
		return nil
	}
	if sug, ok := s.Suggestions[word]; ok {
		return append([]string(nil), sug...)
	}
	// Binary FSA (Java Speller.findReplacements / SuggestEdits Contains probe)
	if s.SuggestFn != nil {
		if out := s.SuggestFn(word); len(out) > 0 {
			return out
		}
	}
	// limited: collect dictionary words within MaxEditDistance (small dicts only)
	if len(s.Words) == 0 || len(s.Words) > 5000 {
		return nil
	}
	var out []string
	for w := range s.Words {
		d := editDistance(word, w)
		// exclude exact dictionary form (Java getSuggestions returns empty for known words)
		if d > 0 && d <= s.MaxEditDistance {
			out = append(out, w)
			if len(out) >= 8 {
				break
			}
		}
	}
	return out
}

func editDistance(a, b string) int {
	// simple Levenshtein on runes
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	prev := make([]int, len(br)+1)
	cur := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			cur[j] = min3(del, ins, sub)
		}
		prev, cur = cur, prev
	}
	return prev[len(br)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
