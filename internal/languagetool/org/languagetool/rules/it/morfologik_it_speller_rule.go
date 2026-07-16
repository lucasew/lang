package it

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikItalianSpellerRuleID = "MORFOLOGIK_RULE_IT"
	ItalianSpellerDict             = "/it/hunspell/it_IT.dict"
)

// MorfologikItalianSpellerRule ports rules.it.MorfologikItalianSpellerRule.
type MorfologikItalianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikItalianSpellerRule() *MorfologikItalianSpellerRule {
	return &MorfologikItalianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikItalianSpellerRuleID, "it", ItalianSpellerDict, nil),
	}
}
