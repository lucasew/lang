package crh

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikCrimeanTatarSpellerRuleID = "MORFOLOGIK_RULE_CRH_UA"
	MorfologikCrimeanTatarSpellerRuleDict = "/crh/hunspell/crh_UA.dict"
)

// MorfologikCrimeanTatarSpellerRule ports rules.crh.MorfologikCrimeanTatarSpellerRule.
type MorfologikCrimeanTatarSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikCrimeanTatarSpellerRule() *MorfologikCrimeanTatarSpellerRule {
	return &MorfologikCrimeanTatarSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikCrimeanTatarSpellerRuleID, "crh", MorfologikCrimeanTatarSpellerRuleDict, nil),
	}
}
