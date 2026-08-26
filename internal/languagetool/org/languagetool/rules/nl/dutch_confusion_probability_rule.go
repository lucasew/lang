package nl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// DutchConfusionProbabilityRule ports org.languagetool.rules.nl.DutchConfusionProbabilityRule.
type DutchConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewDutchConfusionProbabilityRule(lm ngrams.LanguageModel) *DutchConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = false
	return &DutchConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
