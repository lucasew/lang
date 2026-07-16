package zh

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikChineseSpellerRuleID = "MORFOLOGIK_RULE_ZH"
	ChineseSpellerDict = "/zh/hunspell/zh.dict"
)

type MorfologikChineseSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikChineseSpellerRule() *MorfologikChineseSpellerRule {
	return &MorfologikChineseSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikChineseSpellerRuleID, "zh", ChineseSpellerDict, nil),
	}
}
