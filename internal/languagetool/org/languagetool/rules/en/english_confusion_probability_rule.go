package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

// EnglishConfusionProbabilityRule ports
// org.languagetool.rules.en.EnglishConfusionProbabilityRule.
type EnglishConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

const EnglishConfusionRuleID = "EN_CONFUSION_RULE"

func NewEnglishConfusionProbabilityRule(lm ngrams.LanguageModel) *EnglishConfusionProbabilityRule {
	base := ngrams.NewConfusionProbabilityRule(lm, 3)
	base.RuleIDOverride = EnglishConfusionRuleID
	// Java: breaks → brakes
	base.AddExamplePair(
		rules.Wrong("Don't forget to put on the <marker>breaks</marker>."),
		rules.Fixed("Don't forget to put on the <marker>brakes</marker>."),
	)
	return &EnglishConfusionProbabilityRule{ConfusionProbabilityRule: base}
}
