package ca

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
)

// CatalanSpellingMultitokenDict ports
// CatalanMorfologikMultitokenSpeller.SPELLING_MULTITOKEN_DICT_FILENAME.
const CatalanSpellingMultitokenDict = "/ca/ca-ES_spelling_multitoken.dict"

// CatalanMorfologikMultitokenSpeller ports
// org.languagetool.rules.ca.CatalanMorfologikMultitokenSpeller (static getSpeller).
// Optional Factory is for test injection of suggestion functions.
type CatalanMorfologikMultitokenSpeller struct {
	// Factory optional; when set, GetSuggestions uses it instead of the singleton dict.
	Factory MultitokenSpellerFactory
}

// NewCatalanMorfologikMultitokenSpeller constructs a helper with optional factory.
func NewCatalanMorfologikMultitokenSpeller(factory MultitokenSpellerFactory) *CatalanMorfologikMultitokenSpeller {
	return &CatalanMorfologikMultitokenSpeller{Factory: factory}
}

// GetSuggestions returns multitoken spelling suggestions (factory or singleton dict).
func (s *CatalanMorfologikMultitokenSpeller) GetSuggestions(word string) []string {
	if s != nil && s.Factory != nil {
		return GetCatalanMultitokenSpellerSuggestions(s.Factory, word)
	}
	ws := GetWeightedSuggestions(word)
	out := make([]string, 0, len(ws))
	for _, w := range ws {
		out = append(out, w.Word)
	}
	return out
}

var (
	caMultiSpellerOnce sync.Once
	caMultiSpeller     *morfologik.MorfologikSpeller
	// DictExistsFn optional test hook; nil → try NewMorfologikSpeller (dict may be empty/fail-closed).
	caMultiDictExistsFn func(path string) bool
)

// GetSpeller ports CatalanMorfologikMultitokenSpeller.getSpeller (lazy singleton).
// Returns nil when the multitoken dict is unavailable (Java resourceExists check).
func GetSpeller() *morfologik.MorfologikSpeller {
	caMultiSpellerOnce.Do(func() {
		path := CatalanSpellingMultitokenDict
		if caMultiDictExistsFn != nil && !caMultiDictExistsFn(path) {
			return
		}
		// Java: new MorfologikSpeller(path) — max edit distance default for constructor without arg
		// MorfologikSpeller(String) uses Speller.MAX_DISTANCE
		sp := morfologik.NewMorfologikSpeller(path, 1)
		if sp == nil {
			return
		}
		// If dict file never loaded, suggestions stay empty (fail-closed).
		caMultiSpeller = sp
	})
	return caMultiSpeller
}

// ResetCatalanMultitokenSpeller clears the singleton (tests).
func ResetCatalanMultitokenSpeller() {
	caMultiSpellerOnce = sync.Once{}
	caMultiSpeller = nil
}

// SetCatalanMultitokenDictExistsFn sets resourceExists override (tests) and resets singleton.
func SetCatalanMultitokenDictExistsFn(fn func(path string) bool) {
	ResetCatalanMultitokenSpeller()
	caMultiDictExistsFn = fn
}

// GetWeightedSuggestions ports MorfologikSpeller.getSuggestions → WeightedSuggestion list.
func GetWeightedSuggestions(word string) []multitoken.WeightedSuggestion {
	sp := GetSpeller()
	if sp == nil {
		return nil
	}
	// Go MorfologikSpeller returns strings; Java carries dict weights.
	// Use index as weight to preserve relative order (dict weights differ per Java comment).
	raw := sp.GetSuggestions(word)
	out := make([]multitoken.WeightedSuggestion, 0, len(raw))
	for i, s := range raw {
		if s == "" {
			continue
		}
		out = append(out, multitoken.WeightedSuggestion{Word: s, Weight: i})
	}
	return out
}

// MultitokenSpellerFactory returns a spelling suggestions function for a dict path
// (legacy test/helper surface).
type MultitokenSpellerFactory func(dictPath string) (func(word string) []string, error)

// GetCatalanMultitokenSpellerSuggestions returns suggestions via factory, or nil if unavailable.
func GetCatalanMultitokenSpellerSuggestions(factory MultitokenSpellerFactory, word string) []string {
	if factory == nil {
		return nil
	}
	sp, err := factory(CatalanSpellingMultitokenDict)
	if err != nil || sp == nil {
		return nil
	}
	return sp(word)
}
