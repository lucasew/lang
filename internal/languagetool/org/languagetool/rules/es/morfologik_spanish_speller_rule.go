package es

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikSpanishSpellerRuleID = "MORFOLOGIK_RULE_ES"
	SpanishSpellerDict             = "/es/hunspell/es.dict"
)

// MorfologikSpanishSpellerRule ports rules.es.MorfologikSpanishSpellerRule.
type MorfologikSpanishSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikSpanishSpellerRule() *MorfologikSpanishSpellerRule {
	return &MorfologikSpanishSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikSpanishSpellerRuleID, "es", SpanishSpellerDict, nil),
	}
}
