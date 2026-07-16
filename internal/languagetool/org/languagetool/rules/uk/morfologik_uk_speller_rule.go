package uk

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikUkrainianSpellerRuleID = "MORFOLOGIK_RULE_UK_UA"
	UkrainianSpellerDict = "/uk/hunspell/uk_UA.dict"
)

type MorfologikUkrainianSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikUkrainianSpellerRule() *MorfologikUkrainianSpellerRule {
	return &MorfologikUkrainianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikUkrainianSpellerRuleID, "uk", UkrainianSpellerDict, nil),
	}
}
