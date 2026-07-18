package server

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// RuleMeta assigns category and issue type for well-known Java rule ID families.
func RuleMeta(ruleID string) (categoryID, categoryName, issueType, short string) {
	return languagetool.RuleMeta(ruleID)
}

// SoftRuleMeta is a compatibility alias for RuleMeta.
func SoftRuleMeta(ruleID string) (categoryID, categoryName, issueType, short string) {
	return RuleMeta(ruleID)
}

// RuleDescription returns a stable rule-level description for the API.
func RuleDescription(ruleID string) string {
	return languagetool.RuleDescription(ruleID)
}

// SoftRuleDescription is a compatibility alias for RuleDescription.
func SoftRuleDescription(ruleID string) string {
	return RuleDescription(ruleID)
}
