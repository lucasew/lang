package el

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikGreekSpellerRuleID = "MORFOLOGIK_RULE_EL_GR"
	GreekSpellerDict = "/el/hunspell/el_GR.dict"
)

type MorfologikGreekSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikGreekSpellerRule() *MorfologikGreekSpellerRule {
	return &MorfologikGreekSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikGreekSpellerRuleID, "el", GreekSpellerDict, nil),
	}
}
