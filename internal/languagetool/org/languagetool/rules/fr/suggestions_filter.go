package fr

import (
	"regexp"
)

// SuggestionsFilter ports org.languagetool.rules.fr.SuggestionsFilter.
// Drops suggestions that match RemoveSuggestionsRegexp.
type SuggestionsFilter struct{}

func NewSuggestionsFilter() *SuggestionsFilter {
	return &SuggestionsFilter{}
}

// Filter removes suggestions matching the given regex (case-insensitive).
// Invalid regex → keep all (caller should validate).
func (f *SuggestionsFilter) Filter(suggs []string, removeRegexp string) []string {
	p, err := regexp.Compile("(?i)" + removeRegexp)
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
