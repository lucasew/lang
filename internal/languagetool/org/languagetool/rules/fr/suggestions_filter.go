package fr

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SuggestionsFilter ports org.languagetool.rules.fr.SuggestionsFilter.
// Drops suggestions that match RemoveSuggestionsRegexp.
type SuggestionsFilter struct{}

func NewSuggestionsFilter() *SuggestionsFilter {
	return &SuggestionsFilter{}
}

// AcceptRuleMatch ports SuggestionsFilter.acceptRuleMatch.
func (f *SuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	re, ok := arguments["RemoveSuggestionsRegexp"]
	if !ok {
		panic("Missing key 'RemoveSuggestionsRegexp'")
	}
	match.SetSuggestedReplacements(f.Filter(match.GetSuggestedReplacements(), re))
	return match
}

// Filter removes suggestions that fully match the regex (Java Matcher.matches, CASE_INSENSITIVE).
// Invalid regex → keep all (do not invent drops).
func (f *SuggestionsFilter) Filter(suggs []string, removeRegexp string) []string {
	// Java: Pattern.CASE_INSENSITIVE | UNICODE_CASE + Matcher.matches() (full string).
	p, err := regexp.Compile("(?i)^(?:" + removeRegexp + ")$")
	if err != nil {
		return suggs
	}
	var out []string
	for _, s := range suggs {
		if !p.MatchString(s) {
			out = append(out, s)
		}
	}
	return out
}
