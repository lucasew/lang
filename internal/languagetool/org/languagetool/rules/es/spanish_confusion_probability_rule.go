package es

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// SpanishConfusionProbabilityRule ports org.languagetool.rules.es.SpanishConfusionProbabilityRule.
type SpanishConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewSpanishConfusionProbabilityRule(lm ngrams.LanguageModel) *SpanishConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = false
	return &SpanishConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
