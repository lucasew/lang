package it

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// ItalianConfusionProbabilityRule ports org.languagetool.rules.it.ItalianConfusionProbabilityRule.
type ItalianConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewItalianConfusionProbabilityRule(lm ngrams.LanguageModel) *ItalianConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = false
	return &ItalianConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
