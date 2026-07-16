package ar

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// ArabicConfusionProbabilityRule ports org.languagetool.rules.ar.ArabicConfusionProbabilityRule.
type ArabicConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewArabicConfusionProbabilityRule(lm ngrams.LanguageModel) *ArabicConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = false
	return &ArabicConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
