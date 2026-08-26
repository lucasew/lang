package ru

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// RussianConfusionProbabilityRule ports org.languagetool.rules.ru.RussianConfusionProbabilityRule.
type RussianConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewRussianConfusionProbabilityRule(lm ngrams.LanguageModel) *RussianConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = false
	return &RussianConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
