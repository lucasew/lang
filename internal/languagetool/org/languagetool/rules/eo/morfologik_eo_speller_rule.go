package eo

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikEsperantoSpellerRuleID = "MORFOLOGIK_RULE_EO"
	EsperantoSpellerDict = "/eo/hunspell/eo.dict"
)

type MorfologikEsperantoSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikEsperantoSpellerRule() *MorfologikEsperantoSpellerRule {
	return &MorfologikEsperantoSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikEsperantoSpellerRuleID, "eo", EsperantoSpellerDict, nil),
	}
}
