package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CatalanNumberInWordFilter ports org.languagetool.rules.ca.CatalanNumberInWordFilter
// (extends AbstractNumberInWordFilter; isMisspelled/getSuggestions from
// MorfologikCatalanSpellerRule).
type CatalanNumberInWordFilter struct {
	inner *rules.NumberInWordFilter
}

func NewCatalanNumberInWordFilter() *CatalanNumberInWordFilter {
	return &CatalanNumberInWordFilter{
		inner: &rules.NumberInWordFilter{
			// Java: catalanSpellerRule.isMisspelled / getSpellingSuggestions
			IsMisspelled:   FilterDictIsMisspelled,
			GetSuggestions: FilterDictSuggest,
		},
	}
}

// Suggestions ports acceptRuleMatch candidate building (speller-gated; fail-closed without dict).
func (f *CatalanNumberInWordFilter) Suggestions(word string) []string {
	if f == nil || f.inner == nil {
		return nil
	}
	// Without dict: Java always has MorfologikCatalanSpellerRule; fail-closed invent none.
	if !FilterDictAvailable() {
		return nil
	}
	return f.inner.Suggestions(word)
}

// AcceptRuleMatch ports AbstractNumberInWordFilter.acceptRuleMatch.
func (f *CatalanNumberInWordFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
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
