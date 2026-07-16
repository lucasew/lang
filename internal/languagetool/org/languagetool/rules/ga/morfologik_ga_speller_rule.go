package ga

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikIrishSpellerRuleID = "MORFOLOGIK_RULE_GA"
	IrishSpellerDict = "/ga/hunspell/ga_IE.dict"
)

type MorfologikIrishSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikIrishSpellerRule() *MorfologikIrishSpellerRule {
	return &MorfologikIrishSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikIrishSpellerRuleID, "ga", IrishSpellerDict, nil),
	}
}
