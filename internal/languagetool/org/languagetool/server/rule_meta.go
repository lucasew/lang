package server

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// RuleMeta assigns category and issue type for well-known Java rule ID families.
func RuleMeta(ruleID string) (categoryID, categoryName, issueType, short string) {
	return languagetool.RuleMeta(ruleID)
}

// RuleDescription returns a stable rule-level description for the API.
func RuleDescription(ruleID string) string {
	return languagetool.RuleDescription(ruleID)
}
