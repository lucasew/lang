package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// AdverbFilter ports org.languagetool.rules.en.AdverbFilter.
// Maps adverb+noun pattern args to adjective + noun suggestion.
type AdverbFilter struct{}

func NewAdverbFilter() *AdverbFilter {
	return &AdverbFilter{}
}

// Suggest returns "adjective noun" when the adverb maps and differs from the adjective.
// Empty string means leave suggestions unchanged (Java leaves match as-is).
func (f *AdverbFilter) Suggest(adverb, noun string) string {
	adj, ok := adverb2Adj[adverb]
	if !ok || adj == adverb {
		return ""
	}
	return adj + " " + noun
}

// AcceptRuleMatch ports AdverbFilter.acceptRuleMatch.
// Args: adverb, noun — rewrites suggestion when adverb2Adj maps to a different form.
func (f *AdverbFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	adverb := arguments["adverb"]
	noun := arguments["noun"]
	if sug := f.Suggest(adverb, noun); sug != "" {
		match.SetSuggestedReplacement(sug)
	}
	return match
}
