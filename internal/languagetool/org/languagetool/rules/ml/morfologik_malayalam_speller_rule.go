package ml

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikMalayalamSpellerRuleID = "MORFOLOGIK_RULE_ML_IN"
	MorfologikMalayalamSpellerRuleDict = "/ml/hunspell/ml_IN.dict"
)

// MorfologikMalayalamSpellerRule ports rules.ml.MorfologikMalayalamSpellerRule.
type MorfologikMalayalamSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikMalayalamSpellerRule() *MorfologikMalayalamSpellerRule {
	r := &MorfologikMalayalamSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikMalayalamSpellerRuleID, "ml", MorfologikMalayalamSpellerRuleDict, nil),
	}
	// Java MorfologikSpellerRule.initSpeller when binary present.
	r.InitSpellersFromGetters(nil, nil)
	return r
}
