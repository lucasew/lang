package br

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	MorfologikBretonSpellerRuleID   = "MORFOLOGIK_RULE_BR_FR"
	MorfologikBretonSpellerRuleDict = "/br/hunspell/br_FR.dict"
)

// bretonTokenizingPattern ports MorfologikBretonSpellerRule.BRETON_TOKENIZING_CHARS = Pattern.compile("-").
var bretonTokenizingPattern = regexp.MustCompile(`-`)

// MorfologikBretonSpellerRule ports rules.br.MorfologikBretonSpellerRule.
// tokenizingPattern = "-"; setIgnoreTaggedWords. Match uses parent with TokenizingPattern.
type MorfologikBretonSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikBretonSpellerRule() *MorfologikBretonSpellerRule {
	r := &MorfologikBretonSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikBretonSpellerRuleID, "br", MorfologikBretonSpellerRuleDict, nil),
	}
	// Java MorfologikBretonSpellerRule ctor: setIgnoreTaggedWords().
	r.IgnoreTaggedWords = true
	// Java tokenizingPattern(): Pattern.compile("-") — base Match splits getRuleMatches per segment.
	r.TokenizingPattern = bretonTokenizingPattern
	return r
}
