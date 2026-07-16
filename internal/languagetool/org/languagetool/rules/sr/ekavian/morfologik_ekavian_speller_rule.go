package ekavian

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikEkavianSpellerRuleID = "MORFOLOGIK_RULE_SR_EKAVIAN"
	EkavianSpellerDict             = "/sr/dictionary/ekavian/serbian.dict"
)

// MorfologikEkavianSpellerRule ports rules.sr.ekavian.MorfologikEkavianSpellerRule.
type MorfologikEkavianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikEkavianSpellerRule() *MorfologikEkavianSpellerRule {
	return &MorfologikEkavianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikEkavianSpellerRuleID, "sr", EkavianSpellerDict, nil),
	}
}
