package rules

import "sort"

// RuleWithMaxFilter ports org.languagetool.rules.RuleWithMaxFilter.
// Keeps the longest match when a later match is fully included in an earlier
// one from the same AbstractPatternRule (same id + subId).
type RuleWithMaxFilter struct{}

func NewRuleWithMaxFilter() *RuleWithMaxFilter { return &RuleWithMaxFilter{} }

func (f *RuleWithMaxFilter) Filter(ruleMatches []*RuleMatch) []*RuleMatch {
	if len(ruleMatches) == 0 {
		return ruleMatches
	}
	sorted := append([]*RuleMatch(nil), ruleMatches...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].FromPos < sorted[j].FromPos
	})
	var filtered []*RuleMatch
	for i := 0; i < len(sorted); i++ {
		match := sorted[i]
		if i < len(sorted)-1 {
			nextMatch := sorted[i+1]
			for f.Includes(match, nextMatch) && f.haveSameRule(match, nextMatch) && i < len(sorted) {
				i++ // skip next match
				if i < len(sorted)-1 {
					nextMatch = sorted[i+1]
				} else {
					break
				}
			}
		}
		filtered = append(filtered, match)
	}
	return filtered
}

// Includes is package-visible for tests.
func (f *RuleWithMaxFilter) Includes(match, next *RuleMatch) bool {
	return match.FromPos <= next.FromPos && match.ToPos >= next.ToPos
}

func (f *RuleWithMaxFilter) haveSameRule(match, next *RuleMatch) bool {
	pr1, ok1 := match.Rule.(AbstractPatternRule)
	pr2, ok2 := next.Rule.(AbstractPatternRule)
	if !ok1 || !ok2 {
		return false
	}
	id1 := pr1.GetID()
	sub1, sub2 := pr1.GetSubID(), pr2.GetSubID()
	if sub1 == nil && sub2 != nil {
		return false
	}
	if sub1 != nil && sub2 == nil {
		return false
	}
	if id1 == "" || id1 != pr2.GetID() {
		return false
	}
	if sub1 == nil && sub2 == nil {
		return true
	}
	return *sub1 == *sub2
}
