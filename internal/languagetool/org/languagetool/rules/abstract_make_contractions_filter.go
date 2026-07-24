package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// MakeContractionsFilter ports AbstractMakeContractionsFilter: rewrites suggestions
// via a language-specific FixContractions function.
type MakeContractionsFilter struct {
	FixContractions func(suggestion string) string
}

func NewMakeContractionsFilter(fix func(string) string) *MakeContractionsFilter {
	if fix == nil {
		fix = func(s string) string { return s }
	}
	return &MakeContractionsFilter{FixContractions: fix}
}

// AcceptRuleMatch ports AbstractMakeContractionsFilter.acceptRuleMatch.
func (f *MakeContractionsFilter) AcceptRuleMatch(match *RuleMatch, _ map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if match == nil {
		return nil
	}
	match.SetSuggestedReplacements(f.MapSuggestions(match.GetSuggestedReplacements()))
	return match
}

// MapSuggestions applies FixContractions to each suggestion.
func (f *MakeContractionsFilter) MapSuggestions(suggs []string) []string {
	out := make([]string, len(suggs))
	for i, s := range suggs {
		out[i] = f.FixContractions(s)
	}
	return out
}
