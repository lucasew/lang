package rules

import "sort"

// SameRuleGroupFilter ports org.languagetool.rules.SameRuleGroupFilter.
// Keeps only the first match from overlapping matches with the same rule id.
type SameRuleGroupFilter struct{}

func NewSameRuleGroupFilter() *SameRuleGroupFilter { return &SameRuleGroupFilter{} }

func (f *SameRuleGroupFilter) Filter(ruleMatches []*RuleMatch) []*RuleMatch {
	if len(ruleMatches) == 0 {
		return ruleMatches
	}
	// sort by FromPos (RuleMatch.compareTo)
	sorted := append([]*RuleMatch(nil), ruleMatches...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].FromPos < sorted[j].FromPos
	})
	var filtered []*RuleMatch
	for i := 0; i < len(sorted); i++ {
		match := sorted[i]
		for i < len(sorted)-1 && f.overlapAndMatch(match, sorted[i+1]) {
			i++ // skip next match
		}
		filtered = append(filtered, match)
	}
	return filtered
}

func (f *SameRuleGroupFilter) overlapAndMatch(match, next *RuleMatch) bool {
	return f.Overlaps(match, next) && f.haveSameRuleGroup(match, next)
}

// Overlaps is package-visible for tests (Java package-private overlaps).
func (f *SameRuleGroupFilter) Overlaps(match, next *RuleMatch) bool {
	return match.FromPos <= next.ToPos && match.ToPos >= next.FromPos
}

func (f *SameRuleGroupFilter) haveSameRuleGroup(match, next *RuleMatch) bool {
	id1 := ruleIDOf(match.Rule)
	id2 := ruleIDOf(next.Rule)
	return id1 != "" && id1 == id2
}

func ruleIDOf(rule any) string {
	if r, ok := rule.(RuleWithID); ok {
		return r.GetID()
	}
	return ""
}
