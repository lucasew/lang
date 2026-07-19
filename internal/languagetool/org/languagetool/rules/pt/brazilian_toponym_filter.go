package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// BrazilianToponymFilter ports org.languagetool.rules.pt.BrazilianToponymFilter
// (extends RegexRuleFilter — uses regex capture groups 1–3).
type BrazilianToponymFilter struct {
	Map *BrazilianToponymMap
}

func NewBrazilianToponymFilter() *BrazilianToponymFilter {
	return &BrazilianToponymFilter{Map: LoadBrazilianToponymMap()}
}

// AcceptRuleMatch ports RegexRuleFilter.acceptRuleMatch for Brazilian toponyms.
// groups: [0]=full, [1]=toponym, [2]=underlined separator+state region, [3]=state abbrev.
func (f *BrazilianToponymFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string,
	_ *languagetool.AnalyzedSentence, groups []string) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	// Need groups 1..3 (Java matcher.group(1..3)).
	if len(groups) < 4 {
		return nil
	}
	toponym := groups[1]
	underlined := groups[2]
	state := groups[3]
	suggestion := f.Suggest(toponym, underlined, state)
	if suggestion == "" {
		return nil
	}
	match.SetSuggestedReplacement(suggestion)
	return match
}

// Suggest returns the en-dash + state suggestion when the toponym is valid
// and the underlined text is not already that suggestion.
// Empty string means suppress the match.
func (f *BrazilianToponymFilter) Suggest(toponym, underlined, state string) string {
	suggestion := "–" + state
	if suggestion == underlined {
		return ""
	}
	if f.Map == nil || !f.Map.IsValidToponym(toponym) {
		return ""
	}
	return suggestion
}
