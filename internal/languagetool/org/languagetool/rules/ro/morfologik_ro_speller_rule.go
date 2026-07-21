package ro

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikRomanianSpellerRuleID = "MORFOLOGIK_RULE_RO_RO"
	RomanianSpellerDict = "/ro/hunspell/ro_RO.dict"
)

type MorfologikRomanianSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikRomanianSpellerRule() *MorfologikRomanianSpellerRule {
	r := &MorfologikRomanianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikRomanianSpellerRuleID, "ro", RomanianSpellerDict, nil),
	}
	// Java MorfologikSpellerRule.initSpeller when binary present.
	r.InitSpellersFromGetters(nil, nil)
	return r
}
