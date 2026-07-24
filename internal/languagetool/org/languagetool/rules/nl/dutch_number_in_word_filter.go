package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DutchNumberInWordFilter ports org.languagetool.rules.nl.DutchNumberInWordFilter
// (extends AbstractNumberInWordFilter; isMisspelled/getSuggestions from
// MorfologikDutchSpellerRule).
type DutchNumberInWordFilter struct {
	inner *rules.NumberInWordFilter
}

func NewDutchNumberInWordFilter() *DutchNumberInWordFilter {
	return &DutchNumberInWordFilter{
		inner: &rules.NumberInWordFilter{
			// Java: dutchSpellerRule.isMisspelled / getSpellingSuggestions
			IsMisspelled:   FilterDictIsMisspelled,
			GetSuggestions: FilterDictSuggest,
		},
	}
}

// Suggestions ports acceptRuleMatch candidate building (speller-gated; fail-closed without dict).
func (f *DutchNumberInWordFilter) Suggestions(word string) []string {
	if f == nil || f.inner == nil {
		return nil
	}
	// Without dict: Java always has MorfologikDutchSpellerRule; fail-closed invent none.
	if !FilterDictAvailable() {
		return nil
	}
	return f.inner.Suggestions(word)
}

// AcceptRuleMatch ports AbstractNumberInWordFilter.acceptRuleMatch.
func (f *DutchNumberInWordFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
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
