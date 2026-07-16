package zh

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// ChineseConfusionProbabilityRule ports org.languagetool.rules.zh.ChineseConfusionProbabilityRule.
type ChineseConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

func NewChineseConfusionProbabilityRule(lm ngrams.LanguageModel) *ChineseConfusionProbabilityRule {
	r := ngrams.NewConfusionProbabilityRule(lm, 3)
	r.DefaultOff = false
	return &ChineseConfusionProbabilityRule{ConfusionProbabilityRule: r}
}
