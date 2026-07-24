package language

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

func init() {
	languagetool.FilterEnglishRuleMatchesHook = FilterEnglishRuleMatches
}

// FilterEnglishRuleMatches ports English.filterRuleMatches:
//  1. Contraction-aware suggestion whitespace ('… / n't…)
//  2. EN_SIMPLE_REPLACE*GRAMME(S) → locale-violation ITS
//
// Uses LocalMatch.OriginalSurface for errorStr (Java getOriginalErrorStr / setOriginalErrorStr).
// Without surface, contraction prefix logic is skipped (fail-closed; no invent).
func FilterEnglishRuleMatches(matches []languagetool.LocalMatch) []languagetool.LocalMatch {
	if len(matches) == 0 {
		return nil
	}
	out := make([]languagetool.LocalMatch, 0, len(matches))
	for i := range matches {
		m := matches[i]
		errorStr := m.OriginalSurface()
		if len(m.Suggestions) > 0 {
			seen := make(map[string]struct{}, len(m.Suggestions))
			newSugs := make([]string, 0, len(m.Suggestions))
			for _, sug := range m.Suggestions {
				newRepl := sug
				if len(errorStr) > 2 {
					// add a whitespace when the error is in a contraction and the suggestion is not
					if strings.HasPrefix(errorStr, "'") &&
						!strings.HasPrefix(newRepl, "'") &&
						!strings.HasPrefix(newRepl, "’") &&
						!strings.HasPrefix(newRepl, " ") {
						newRepl = " " + newRepl
					}
					if strings.HasPrefix(errorStr, "n't") &&
						!strings.HasPrefix(newRepl, "n't") &&
						!strings.HasPrefix(newRepl, "n’t") {
						newRepl = " " + newRepl
					}
				}
				if _, ok := seen[newRepl]; !ok {
					seen[newRepl] = struct{}{}
					newSugs = append(newSugs, newRepl)
				}
			}
			m.Suggestions = newSugs
		}
		// Java: getSpecificRuleId starts with EN_SIMPLE_REPLACE and ends with GRAMME(S)
		id := m.RuleID
		if strings.HasPrefix(id, "EN_SIMPLE_REPLACE") &&
			(strings.HasSuffix(id, "GRAMME") || strings.HasSuffix(id, "GRAMMES")) {
			m.IssueType = "locale-violation"
		}
		out = append(out, m)
	}
	return out
}
