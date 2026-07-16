package rules

import "strings"

// MatchSpan is a simple from/to range for overlap checks.
type MatchSpan struct {
	From, To int
}

// SuppressIfAnyRuleMatchesFilter ports org.languagetool.rules.SuppressIfAnyRuleMatchesFilter.
// MatchesInSentence reports rule matches for a given rule ID in a rewritten sentence.
type SuppressIfAnyRuleMatchesFilter struct {
	// MatchesInSentence returns match spans for ruleID in newSentence.
	MatchesInSentence func(ruleID, newSentence string) []MatchSpan
}

func NewSuppressIfAnyRuleMatchesFilter(fn func(ruleID, newSentence string) []MatchSpan) *SuppressIfAnyRuleMatchesFilter {
	return &SuppressIfAnyRuleMatchesFilter{MatchesInSentence: fn}
}

// ShouldSuppress is true if any replacement creates an overlapping match for any ruleIDs.
func (f *SuppressIfAnyRuleMatchesFilter) ShouldSuppress(sentence string, fromPos, toPos int, replacements []string, ruleIDsCSV string) bool {
	if f.MatchesInSentence == nil {
		return false
	}
	ids := strings.Split(ruleIDsCSV, ",")
	for _, replacement := range replacements {
		if fromPos < 0 || toPos > len(sentence) || fromPos > toPos {
			continue
		}
		newSentence := sentence[:fromPos] + replacement + sentence[toPos:]
		for _, id := range ids {
			id = strings.TrimSpace(id)
			for _, m := range f.MatchesInSentence(id, newSentence) {
				// overlap with original match range (Java logic)
				if (m.To >= fromPos && m.To <= toPos) || (toPos >= m.From && toPos <= m.To) {
					return true
				}
			}
		}
	}
	return false
}
