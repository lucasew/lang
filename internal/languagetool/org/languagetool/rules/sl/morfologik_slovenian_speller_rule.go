package sl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikSlovenianSpellerRuleID = "MORFOLOGIK_RULE_SL"
	SlovenianSpellerDict             = "/sl/hunspell/sl_SI.dict"
)

// MorfologikSlovenianSpellerRule ports rules.sl.MorfologikSlovenianSpellerRule.
type MorfologikSlovenianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikSlovenianSpellerRule() *MorfologikSlovenianSpellerRule {
	return &MorfologikSlovenianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikSlovenianSpellerRuleID, "sl", SlovenianSpellerDict, nil),
	}
}
