package sl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikSlovenianSpellerRuleID ports MorfologikSlovenianSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_SL_SI" (not MORFOLOGIK_RULE_SL).
	MorfologikSlovenianSpellerRuleID = "MORFOLOGIK_RULE_SL_SI"
	// SlovenianSpellerDict ports MorfologikSlovenianSpellerRule.getFileName() → RESOURCE_FILENAME.
	// Java: "/sl/hunspell/sl_SI.dict"
	SlovenianSpellerDict = "/sl/hunspell/sl_SI.dict"
)

// MorfologikSlovenianSpellerRule ports rules.sl.MorfologikSlovenianSpellerRule.
type MorfologikSlovenianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikSlovenianSpellerRule() *MorfologikSlovenianSpellerRule {
	r := &MorfologikSlovenianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikSlovenianSpellerRuleID, "sl", SlovenianSpellerDict, nil),
	}
	// Java MorfologikSpellerRule.initSpeller when binary present.
	r.InitSpellersFromGetters(nil, nil)
	return r
}
