package fr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// FrenchConfusionProbabilityRule ports org.languagetool.rules.fr.FrenchConfusionProbabilityRule.
type FrenchConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewFrenchConfusionProbabilityRule(lm ngrams.LanguageModel) *FrenchConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = false
	return &FrenchConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
