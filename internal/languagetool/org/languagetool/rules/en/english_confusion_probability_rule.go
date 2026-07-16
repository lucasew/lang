package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// EnglishConfusionProbabilityRule ports
// org.languagetool.rules.en.EnglishConfusionProbabilityRule.
type EnglishConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

const EnglishConfusionRuleID = "EN_CONFUSION_RULE"

func NewEnglishConfusionProbabilityRule(lm ngrams.LanguageModel) *EnglishConfusionProbabilityRule {
	base := ngrams.NewConfusionProbabilityRule(lm, 3)
	base.RuleIDOverride = EnglishConfusionRuleID
	return &EnglishConfusionProbabilityRule{ConfusionProbabilityRule: base}
}
