package rules

// RuleMatchFilter ports org.languagetool.rules.RuleMatchFilter.
// AnnotatedText is represented as plain text on this surface port.
type RuleMatchFilter interface {
	// Filter transforms/filters matches for the given text.
	Filter(ruleMatches []*RuleMatch, text string) []*RuleMatch
}

// RuleMatchFilterFunc adapts a function to RuleMatchFilter.
type RuleMatchFilterFunc func(ruleMatches []*RuleMatch, text string) []*RuleMatch

func (f RuleMatchFilterFunc) Filter(ruleMatches []*RuleMatch, text string) []*RuleMatch {
	return f(ruleMatches, text)
}

// IdentityRuleMatchFilter returns matches unchanged.
func IdentityRuleMatchFilter() RuleMatchFilter {
	return RuleMatchFilterFunc(func(m []*RuleMatch, _ string) []*RuleMatch { return m })
}
