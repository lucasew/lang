package br

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikBretonSpellerRuleID = "MORFOLOGIK_RULE_BR_FR"
	MorfologikBretonSpellerRuleDict = "/br/hunspell/br_FR.dict"
)

// MorfologikBretonSpellerRule ports rules.br.MorfologikBretonSpellerRule.
type MorfologikBretonSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikBretonSpellerRule() *MorfologikBretonSpellerRule {
	return &MorfologikBretonSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikBretonSpellerRuleID, "br", MorfologikBretonSpellerRuleDict, nil),
	}
}
