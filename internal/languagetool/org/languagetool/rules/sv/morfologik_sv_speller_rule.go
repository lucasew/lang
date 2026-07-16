package sv

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikSwedishSpellerRuleID = "MORFOLOGIK_RULE_SV_SE"
	SwedishSpellerDict = "/sv/hunspell/sv_SE.dict"
)

type MorfologikSwedishSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikSwedishSpellerRule() *MorfologikSwedishSpellerRule {
	return &MorfologikSwedishSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikSwedishSpellerRuleID, "sv", SwedishSpellerDict, nil),
	}
}
