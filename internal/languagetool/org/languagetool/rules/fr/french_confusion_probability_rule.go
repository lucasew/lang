package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

// FrenchConfusionProbabilityRule ports org.languagetool.rules.fr.FrenchConfusionProbabilityRule.
type FrenchConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewFrenchConfusionProbabilityRule(lm ngrams.LanguageModel) *FrenchConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = false
	// Java: pris → prix (trailing unclosed <marker> kept as upstream)
	r.AddExamplePair(
		rules.Wrong("Friedman résume cela en écrivant que le système de <marker>pris</marker> libres remplit trois fonctions.<marker>"),
		rules.Fixed("Friedman résume cela en écrivant que le système de <marker>prix</marker> libres remplit trois fonctions.<marker>"),
	)
	return &FrenchConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
