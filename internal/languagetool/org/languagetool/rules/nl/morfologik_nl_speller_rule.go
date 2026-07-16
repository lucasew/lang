package nl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikDutchSpellerRuleID = "MORFOLOGIK_RULE_NL"
	DutchSpellerDict             = "/nl/hunspell/nl_NL.dict"
)

// MorfologikDutchSpellerRule ports rules.nl.MorfologikDutchSpellerRule.
type MorfologikDutchSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikDutchSpellerRule() *MorfologikDutchSpellerRule {
	return &MorfologikDutchSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikDutchSpellerRuleID, "nl", DutchSpellerDict, nil),
	}
}
