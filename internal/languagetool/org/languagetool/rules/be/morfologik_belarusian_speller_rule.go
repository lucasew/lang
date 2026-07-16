package be

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikBelarusianSpellerRuleID = "MORFOLOGIK_RULE_BE_BY"
	MorfologikBelarusianSpellerRuleDict = "/be/hunspell/be_BY.dict"
)

// MorfologikBelarusianSpellerRule ports rules.be.MorfologikBelarusianSpellerRule.
type MorfologikBelarusianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikBelarusianSpellerRule() *MorfologikBelarusianSpellerRule {
	return &MorfologikBelarusianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikBelarusianSpellerRuleID, "be", MorfologikBelarusianSpellerRuleDict, nil),
	}
}
