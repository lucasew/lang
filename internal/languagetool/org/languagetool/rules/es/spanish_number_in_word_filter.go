package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SpanishNumberInWordFilter ports org.languagetool.rules.es.SpanishNumberInWordFilter
// (extends AbstractNumberInWordFilter; isMisspelled/getSuggestions from
// MorfologikSpanishSpellerRule).
type SpanishNumberInWordFilter struct {
	inner *rules.NumberInWordFilter
}

func NewSpanishNumberInWordFilter() *SpanishNumberInWordFilter {
	return &SpanishNumberInWordFilter{
		inner: &rules.NumberInWordFilter{
			// Java: spanishSpellerRule.isMisspelled / getSpellingSuggestions
			IsMisspelled:   FilterDictIsMisspelled,
			GetSuggestions: FilterDictSuggest,
		},
	}
}

// Suggestions ports acceptRuleMatch candidate building (speller-gated; fail-closed without dict).
func (f *SpanishNumberInWordFilter) Suggestions(word string) []string {
	if f == nil || f.inner == nil {
		return nil
	}
	// Without dict: Java always has MorfologikSpanishSpellerRule; fail-closed invent none.
	if !FilterDictAvailable() {
		return nil
	}
	return f.inner.Suggestions(word)
}

// AcceptRuleMatch ports AbstractNumberInWordFilter.acceptRuleMatch.
func (f *SpanishNumberInWordFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
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
