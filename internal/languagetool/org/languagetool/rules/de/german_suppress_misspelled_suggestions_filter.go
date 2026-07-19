package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanSuppressMisspelledSuggestionsFilter ports
// org.languagetool.rules.de.GermanSuppressMisspelledSuggestionsFilter
// (empty subclass of AbstractSuppressMisspelledSuggestionsFilter).
//
// Java: language.getDefaultSpellingRule(); null rule → isMisspelled false (keep all).
// Default IsMisspelled is FilterDictIsMisspelled (false when no WireGermanFilterSpeller).
type GermanSuppressMisspelledSuggestionsFilter struct {
	*rules.AbstractSuppressMisspelledSuggestionsFilter
}

func NewGermanSuppressMisspelledSuggestionsFilter() *GermanSuppressMisspelledSuggestionsFilter {
	return &GermanSuppressMisspelledSuggestionsFilter{
		AbstractSuppressMisspelledSuggestionsFilter: &rules.AbstractSuppressMisspelledSuggestionsFilter{
			// Java: GermanyGerman.getDefaultSpellingRule().isMisspelled (via abstract)
			IsMisspelled: FilterDictIsMisspelled,
		},
	}
}

// SetIsMisspelled wires the speller hook (tests / host override).
func (f *GermanSuppressMisspelledSuggestionsFilter) SetIsMisspelled(fn func(string) bool) {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return
	}
	f.IsMisspelled = fn
}

// FilterSuggestions is a test helper matching prior surface API:
// drops misspelled suggestions; empty + suppressMatch → drop match.
func (f *GermanSuppressMisspelledSuggestionsFilter) FilterSuggestions(suggs []string, suppressMatch bool) (kept []string, keepMatch bool) {
	var miss rules.MisspelledFunc
	if f != nil && f.AbstractSuppressMisspelledSuggestionsFilter != nil {
		miss = f.IsMisspelled
	}
	if miss == nil {
		// Java: null SpellingCheckRule → isMisspelled false for all → keep all
		return suggs, true
	}
	for _, s := range suggs {
		if !miss(s) {
			kept = append(kept, s)
		}
	}
	if len(kept) == 0 && suppressMatch {
		return nil, false
	}
	return kept, true
}

// AcceptRuleMatch ports AbstractSuppressMisspelledSuggestionsFilter.acceptRuleMatch.
func (f *GermanSuppressMisspelledSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return nil
	}
	return f.AbstractSuppressMisspelledSuggestionsFilter.AcceptRuleMatch(match, arguments)
}
