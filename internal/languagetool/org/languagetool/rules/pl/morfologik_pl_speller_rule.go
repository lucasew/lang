package pl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikPolishSpellerRuleID = "MORFOLOGIK_RULE_PL"
	PolishSpellerDict             = "/pl/hunspell/pl_PL.dict"
)

// MorfologikPolishSpellerRule ports rules.pl.MorfologikPolishSpellerRule.
type MorfologikPolishSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikPolishSpellerRule() *MorfologikPolishSpellerRule {
	return &MorfologikPolishSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikPolishSpellerRuleID, "pl", PolishSpellerDict, nil),
	}
}
