package tl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikTagalogSpellerRuleID   = "MORFOLOGIK_RULE_TL"
	MorfologikTagalogSpellerRuleDict = "/tl/hunspell/tl_PH.dict"
)

// MorfologikTagalogSpellerRule ports language.tl.MorfologikTagalogSpellerRule.
type MorfologikTagalogSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikTagalogSpellerRule() *MorfologikTagalogSpellerRule {
	return &MorfologikTagalogSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikTagalogSpellerRuleID, "tl", MorfologikTagalogSpellerRuleDict, nil),
	}
}
