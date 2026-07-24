package lt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikLithuanianSpellerRuleID = "MORFOLOGIK_RULE_LT_LT"
	MorfologikLithuanianSpellerRuleDict = "/lt/hunspell/lt_LT.dict"
)

// MorfologikLithuanianSpellerRule ports rules.lt.MorfologikLithuanianSpellerRule.
type MorfologikLithuanianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikLithuanianSpellerRule() *MorfologikLithuanianSpellerRule {
	r := &MorfologikLithuanianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikLithuanianSpellerRuleID, "lt", MorfologikLithuanianSpellerRuleDict, nil),
	}
	// Java MorfologikSpellerRule.initSpeller when binary present.
	r.InitSpellersFromGetters(nil, nil)
	return r
}
