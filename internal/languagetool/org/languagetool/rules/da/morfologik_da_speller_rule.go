package da

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikDanishSpellerRuleID = "MORFOLOGIK_RULE_DA_DK"
	DanishSpellerDict = "/da/hunspell/da_DK.dict"
)

type MorfologikDanishSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikDanishSpellerRule() *MorfologikDanishSpellerRule {
	return &MorfologikDanishSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikDanishSpellerRuleID, "da", DanishSpellerDict, nil),
	}
}
