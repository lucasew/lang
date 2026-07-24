package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

// SpanishConfusionProbabilityRule ports org.languagetool.rules.es.SpanishConfusionProbabilityRule.
type SpanishConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewSpanishConfusionProbabilityRule(lm ngrams.LanguageModel) *SpanishConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = false
	// Java: tubo → tuvo
	r.AddExamplePair(
		rules.Wrong("El proyecto no <marker>tubo</marker> una buena acogida."),
		rules.Fixed("El proyecto no <marker>tuvo</marker> una buena acogida."),
	)
	return &SpanishConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
