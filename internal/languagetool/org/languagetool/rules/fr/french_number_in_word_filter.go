package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FrenchNumberInWordFilter ports org.languagetool.rules.fr.FrenchNumberInWordFilter
// (extends AbstractNumberInWordFilter; isMisspelled/getSuggestions from
// MorfologikFrenchSpellerRule / French.getDefaultSpellingRule()).
type FrenchNumberInWordFilter struct {
	inner *rules.NumberInWordFilter
}

func NewFrenchNumberInWordFilter() *FrenchNumberInWordFilter {
	return &FrenchNumberInWordFilter{
		inner: &rules.NumberInWordFilter{
			// Java: frenchSpellerRule.isMisspelled / getSpellingSuggestions
			IsMisspelled:   FilterDictIsMisspelled,
			GetSuggestions: FilterDictSuggest,
		},
	}
}

// Suggestions ports acceptRuleMatch candidate building (speller-gated; fail-closed without dict).
func (f *FrenchNumberInWordFilter) Suggestions(word string) []string {
	if f == nil || f.inner == nil {
		return nil
	}
	// Without dict: Java always has default spelling rule; fail-closed invent none.
	if !FilterDictAvailable() {
		return nil
	}
	return f.inner.Suggestions(word)
}

// AcceptRuleMatch ports AbstractNumberInWordFilter.acceptRuleMatch.
func (f *FrenchNumberInWordFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if !FilterDictAvailable() {
		return nil
	}
	if f.inner == nil {
		return nil
	}
	return f.inner.AcceptRuleMatch(match, arguments, 0, nil, nil)
}
