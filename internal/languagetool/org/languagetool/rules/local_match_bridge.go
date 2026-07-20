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
		fr := NewFakeRule(id)
		// Preserve Rule.getTags so JSON rule.tags / CLI Tags survive the bridge.
		if len(m.Tags) > 0 {
			tags := make([]Tag, len(m.Tags))
			for i, t := range m.Tags {
				tags[i] = Tag(t)
			}
			fr.SetTags(tags)
		} else if m.IsPicky {
			fr.SetTags([]Tag{TagPicky})
		}
		if m.TempOff {
			fr.SetDefaultTempOff()
		}
		rm := NewRuleMatch(fr, sentence, m.FromPos, m.ToPos, m.Message)
		if len(m.Suggestions) > 0 {
			rm.SetSuggestedReplacements(m.Suggestions)
		}
		if m.ShortMessage != "" {
			rm.ShortMessage = m.ShortMessage
		}
		rm.IssueType = m.IssueType
		rm.CategoryID = m.CategoryID
		rm.CategoryName = m.CategoryName
		out = append(out, rm)
	}
	return out
}
