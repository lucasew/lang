package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// FromLocalMatches builds RuleMatch list for CLI/print/JSON from LocalMatch.
// sentence is attached for context (often AnalyzePlain of the full text).
func FromLocalMatches(ms []languagetool.LocalMatch, sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	if len(ms) == 0 {
		return nil
	}
	out := make([]*RuleMatch, 0, len(ms))
	for _, m := range ms {
		id := m.RuleID
		if id == "" {
			id = "LOCAL_MATCH"
		}
		rm := NewRuleMatch(NewFakeRule(id), sentence, m.FromPos, m.ToPos, m.Message)
		if len(m.Suggestions) > 0 {
			rm.SetSuggestedReplacements(m.Suggestions)
		}
		out = append(out, rm)
	}
	return out
}
