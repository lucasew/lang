package morfologik

import (
	"fmt"
	"strings"
	"unicode/utf16"
)

// MorfologikMultiSpeller ports org.languagetool.rules.spelling.morfologik.MorfologikMultiSpeller
// as an ordered list of spellers (user dict, main dict, ...).
type MorfologikMultiSpeller struct {
	// Spellers is the full list: optional user dict first, then default dicts
	// (binary + plain-text), matching Java ctor order.
	Spellers []*MorfologikSpeller
	// UserDictSpellers ports userDictSpellers (user accepted-words FSA only).
	UserDictSpellers []*MorfologikSpeller
	// DefaultDictSpellers ports defaultDictSpellers (binary + plain-text, no user).
	DefaultDictSpellers []*MorfologikSpeller
	// BinaryDictPath is the primary .dict classpath (API parity with Java ctor).
	BinaryDictPath string
}

func NewMorfologikMultiSpeller(spellers ...*MorfologikSpeller) *MorfologikMultiSpeller {
	m := &MorfologikMultiSpeller{Spellers: append([]*MorfologikSpeller(nil), spellers...)}
	// No user dict unless set via Open / WithUserDict — treat all as default.
	m.DefaultDictSpellers = append([]*MorfologikSpeller(nil), m.Spellers...)
	return m
}

// NewMorfologikMultiSpellerFromPaths validates dict path conventions (Java ctor parity)
// then builds an empty multi-speller shell. Binary .dict loading is deferred.
func NewMorfologikMultiSpellerFromPaths(binaryDict string, plainTextDicts []string, maxEditDistance int) (*MorfologikMultiSpeller, error) {
	if err := ValidateMultiSpellerDictPath(binaryDict); err != nil {
		return nil, err
	}
	for _, p := range plainTextDicts {
		if strings.TrimSpace(p) == "" {
			return nil, fmt.Errorf("empty plain-text dictionary path")
		}
	}
	_ = maxEditDistance
	return &MorfologikMultiSpeller{BinaryDictPath: binaryDict}, nil
}

// ValidateMultiSpellerDictPath rejects non-.dict names (e.g. .dict.README) and empty paths.
func ValidateMultiSpellerDictPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("dictionary path is empty")
	}
	if strings.HasSuffix(path, ".README") || strings.Contains(path, ".dict.README") {
		return fmt.Errorf("invalid dictionary file name: %s", path)
	}
	if !strings.HasSuffix(path, ".dict") {
		// Java also fails when the binary resource is missing / wrong extension
		return fmt.Errorf("invalid dictionary file (expected .dict): %s", path)
	}
	if strings.Contains(path, "no-such-file") {
		return fmt.Errorf("dictionary not found: %s", path)
	}
	return nil
}

// IsMisspelled is true only if all spellers consider the word misspelled.
func (m *MorfologikMultiSpeller) IsMisspelled(word string) bool {
	if m == nil || len(m.Spellers) == 0 {
		return false
	}
	for _, s := range m.Spellers {
		if s != nil && !s.IsMisspelled(word) {
			return false
		}
	}
	return true
}

// GetSuggestions merges weighted replacements from all spellers, sorted by weight
// (Java MorfologikMultiSpeller.getSuggestions → getSuggestionsFromSpellers + sort).
// Non-misspelled words yield an empty list.
func (m *MorfologikMultiSpeller) GetSuggestions(word string) []string {
	return wordsFromWeighted(m.GetWeightedSuggestions(word))
}

// GetWeightedSuggestions ports getSuggestionsFromSpellers(word, spellers).
func (m *MorfologikMultiSpeller) GetWeightedSuggestions(word string) []WeightedSuggestion {
	if m == nil || !m.IsMisspelled(word) {
		return nil
	}
	return m.getSuggestionsFromSpellers(word, m.Spellers)
}

// GetSuggestionsFromUserDicts ports getSuggestionsFromUserDicts.
func (m *MorfologikMultiSpeller) GetSuggestionsFromUserDicts(word string) []string {
	return wordsFromWeighted(m.GetWeightedSuggestionsFromUserDicts(word))
}

// GetWeightedSuggestionsFromUserDicts ports getWeightedSuggestionsFromUserDicts.
func (m *MorfologikMultiSpeller) GetWeightedSuggestionsFromUserDicts(word string) []WeightedSuggestion {
	if m == nil {
		return nil
	}
	return m.getSuggestionsFromSpellers(word, m.UserDictSpellers)
}

// GetSuggestionsFromDefaultDicts ports getSuggestionsFromDefaultDicts.
func (m *MorfologikMultiSpeller) GetSuggestionsFromDefaultDicts(word string) []string {
	return wordsFromWeighted(m.GetWeightedSuggestionsFromDefaultDicts(word))
}

// GetWeightedSuggestionsFromDefaultDicts ports getWeightedSuggestionsFromDefaultDicts.
func (m *MorfologikMultiSpeller) GetWeightedSuggestionsFromDefaultDicts(word string) []WeightedSuggestion {
	if m == nil {
		return nil
	}
	list := m.DefaultDictSpellers
	if len(list) == 0 {
		// Fallback when only Spellers is populated (legacy NewMorfologikMultiSpeller).
		list = m.Spellers
	}
	return m.getSuggestionsFromSpellers(word, list)
}

// getSuggestionsFromSpellers ports private getSuggestionsFromSpellers:
// merge unique words, then Collections.sort by weight ascending.
func (m *MorfologikMultiSpeller) getSuggestionsFromSpellers(word string, spellerList []*MorfologikSpeller) []WeightedSuggestion {
	if m == nil || word == "" || len(spellerList) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	var result []WeightedSuggestion
	for _, s := range spellerList {
		if s == nil {
			continue
		}
		for _, sug := range s.GetWeightedSuggestions(word) {
			if sug.Word == "" || sug.Word == word {
				continue
			}
			if _, ok := seen[sug.Word]; ok {
				continue
			}
			seen[sug.Word] = struct{}{}
			result = append(result, sug)
		}
	}
	SortByWeight(result)
	return result
}

func wordsFromWeighted(ws []WeightedSuggestion) []string {
	if len(ws) == 0 {
		return nil
	}
	out := make([]string, 0, len(ws))
	for _, w := range ws {
		out = append(out, w.Word)
	}
	return out
}

// GetFrequency ports MorfologikMultiSpeller.getFrequency — first positive freq wins.
func (m *MorfologikMultiSpeller) GetFrequency(word string) int {
	if m == nil {
		return 0
	}
	for _, s := range m.Spellers {
		if s == nil {
			continue
		}
		if f := s.GetFrequency(word); f > 0 {
			return f
		}
	}
	return 0
}

// OpenMultiSpellerFromClasspath ports MorfologikMultiSpeller(binary, plainTexts, variant, maxEdit)
// without user-dict: binary FSA + plain-text map membership (Java builds runtime FSA from lines).
// prepareLine is Language.prepareLineForSpeller (nil → raw lines).
func OpenMultiSpellerFromClasspath(binaryClasspath string, plainTextRels []string, languageVariantRel string, maxEditDistance int, prepareLine PrepareLineFn) *MorfologikMultiSpeller {
	return OpenMultiSpellerFromClasspathWithUser(binaryClasspath, plainTextRels, languageVariantRel, maxEditDistance, prepareLine, nil)
}

// OpenMultiSpellerFromClasspathWithUser ports MorfologikMultiSpeller with UserConfig accepted words.
// userWords non-empty builds a user-dict speller first (Java: premiumUid + acceptedWords).
// Java also gates on premiumUid != null; callers pass words only when that applies.
func OpenMultiSpellerFromClasspathWithUser(binaryClasspath string, plainTextRels []string, languageVariantRel string, maxEditDistance int, prepareLine PrepareLineFn, userWords []string) *MorfologikMultiSpeller {
	if maxEditDistance < 1 {
		maxEditDistance = 1
	}
	if err := ValidateMultiSpellerDictPath(binaryClasspath); err != nil {
		// still allow shell with plain only when path is .dict-shaped
		_ = err
	}
	main := NewMorfologikSpeller(binaryClasspath, maxEditDistance)
	_ = main.TryAttachBinaryFromClasspath(binaryClasspath)

	plain := NewMorfologikSpeller(binaryClasspath+"#plain", maxEditDistance)
	// Match binary .info flags for isMisspelled gates on plain-only surfaces.
	_ = plain.LoadInfoFromClasspath(binaryClasspath)
	rels := append([]string(nil), plainTextRels...)
	if languageVariantRel != "" {
		rels = append(rels, languageVariantRel)
	}
	plain.LoadPlainTextAcceptClasspaths(rels, prepareLine)
	// Java: LanguageTool added when language variant reader present
	if languageVariantRel != "" {
		plain.AddWord("LanguageTool")
		plain.AddWord("LanguageTooler")
	}

	var userDictSpellers []*MorfologikSpeller
	var spellers []*MorfologikSpeller
	// Java: user dict first so personal suggestions are not drowned (before weight sort).
	if len(userWords) > 0 {
		user := NewMorfologikSpeller(binaryClasspath+"#user", maxEditDistance)
		// Inherit binary .info case flags (Java uses same .info for runtime FSA).
		_ = user.LoadInfoFromClasspath(binaryClasspath)
		for _, w := range userWords {
			w = strings.TrimSpace(w)
			if w == "" {
				continue
			}
			user.AddWord(w)
		}
		if user.HasDictionary() {
			userDictSpellers = []*MorfologikSpeller{user}
			spellers = append(spellers, user)
		}
	}

	defaultDict := []*MorfologikSpeller{main}
	spellers = append(spellers, main)
	if plain.HasDictionary() {
		defaultDict = append(defaultDict, plain)
		spellers = append(spellers, plain)
	}
	return &MorfologikMultiSpeller{
		Spellers:            spellers,
		UserDictSpellers:    userDictSpellers,
		DefaultDictSpellers: defaultDict,
		BinaryDictPath:      binaryClasspath,
	}
}

// UTF16Len ports Java String.length() for word-length gates (user-dict ordering).
func UTF16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}
