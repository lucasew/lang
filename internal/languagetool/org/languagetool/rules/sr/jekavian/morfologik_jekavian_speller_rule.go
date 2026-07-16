package jekavian

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikJekavianSpellerRuleID = "MORFOLOGIK_RULE_SR_JEKAVIAN"
	JekavianSpellerDict             = "/sr/dictionary/jekavian/serbian.dict"
)

// MorfologikJekavianSpellerRule ports rules.sr.jekavian.MorfologikJekavianSpellerRule.
type MorfologikJekavianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikJekavianSpellerRule() *MorfologikJekavianSpellerRule {
	return &MorfologikJekavianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikJekavianSpellerRuleID, "sr", JekavianSpellerDict, nil),
	}
}
