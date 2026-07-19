package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DutchSuppressMisspelledSuggestionsFilter ports
// org.languagetool.rules.nl.DutchSuppressMisspelledSuggestionsFilter
// (empty subclass of AbstractSuppressMisspelledSuggestionsFilter).
//
// Java: language.getDefaultSpellingRule(); null rule → isMisspelled false (keep all).
// Default IsMisspelled is FilterDictIsMisspelled (false when no WireDutchFilterSpeller).
type DutchSuppressMisspelledSuggestionsFilter struct {
	*rules.AbstractSuppressMisspelledSuggestionsFilter
}

func NewDutchSuppressMisspelledSuggestionsFilter() *DutchSuppressMisspelledSuggestionsFilter {
	return &DutchSuppressMisspelledSuggestionsFilter{
		AbstractSuppressMisspelledSuggestionsFilter: &rules.AbstractSuppressMisspelledSuggestionsFilter{
			// Java: Dutch default spelling rule via abstract isMisspelled
			IsMisspelled: FilterDictIsMisspelled,
		},
	}
}

// SetIsMisspelled wires the speller hook (tests / host override).
func (f *DutchSuppressMisspelledSuggestionsFilter) SetIsMisspelled(fn func(string) bool) {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return
	}
	f.IsMisspelled = fn
}

// FilterSuggestions is a test helper matching prior surface API.
func (f *DutchSuppressMisspelledSuggestionsFilter) FilterSuggestions(suggs []string, suppressMatch bool) (kept []string, keepMatch bool) {
	var miss rules.MisspelledFunc
	if f != nil && f.AbstractSuppressMisspelledSuggestionsFilter != nil {
		miss = f.IsMisspelled
	}
	if miss == nil {
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
func (f *DutchSuppressMisspelledSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return nil
	}
	return f.AbstractSuppressMisspelledSuggestionsFilter.AcceptRuleMatch(match, arguments)
}
