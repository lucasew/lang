package fa

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikPersianSpellerRuleID = "MORFOLOGIK_RULE_FA"
	PersianSpellerDict = "/fa/hunspell/fa.dict"
)

type MorfologikPersianSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikPersianSpellerRule() *MorfologikPersianSpellerRule {
	return &MorfologikPersianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikPersianSpellerRuleID, "fa", PersianSpellerDict, nil),
	}
}
