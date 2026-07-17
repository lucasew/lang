package server

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// SoftRuleMeta assigns category and issue type for well-known rule ID families.
func SoftRuleMeta(ruleID string) (categoryID, categoryName, issueType, short string) {
	return languagetool.SoftRuleMeta(ruleID)
}

// SoftRuleDescription returns a stable rule-level description for the API.
func SoftRuleDescription(ruleID string) string {
	return languagetool.SoftRuleDescription(ruleID)
}
