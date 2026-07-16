package ca

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikCatalanSpellerRuleID = "MORFOLOGIK_RULE_CA"
	CatalanSpellerDict             = "/ca/hunspell/ca.dict"
)

// MorfologikCatalanSpellerRule ports rules.ca.MorfologikCatalanSpellerRule.
type MorfologikCatalanSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikCatalanSpellerRule() *MorfologikCatalanSpellerRule {
	return &MorfologikCatalanSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikCatalanSpellerRuleID, "ca", CatalanSpellerDict, nil),
	}
}
