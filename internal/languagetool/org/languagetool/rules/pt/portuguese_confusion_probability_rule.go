package pt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// PortugueseConfusionProbabilityRule ports org.languagetool.rules.pt.PortugueseConfusionProbabilityRule.
type PortugueseConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewPortugueseConfusionProbabilityRule(lm ngrams.LanguageModel) *PortugueseConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = true
	return &PortugueseConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
