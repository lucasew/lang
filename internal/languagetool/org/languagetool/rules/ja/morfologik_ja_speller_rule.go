package ja

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikJapaneseSpellerRuleID = "MORFOLOGIK_RULE_JA"
	JapaneseSpellerDict = "/ja/hunspell/ja.dict"
)

type MorfologikJapaneseSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikJapaneseSpellerRule() *MorfologikJapaneseSpellerRule {
	return &MorfologikJapaneseSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikJapaneseSpellerRuleID, "ja", JapaneseSpellerDict, nil),
	}
}
