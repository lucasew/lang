package ngrams

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// ConfusionRuleID ports deprecated ConfusionProbabilityRule.RULE_ID.
const ConfusionRuleID = "CONFUSION_RULE"

// MinCoverage ports ConfusionProbabilityRule.MIN_COVERAGE.
const MinCoverage = 0.5

// ConfusionProbabilityRule is a metadata + data-holder port of
// org.languagetool.rules.ngrams.ConfusionProbabilityRule.
type ConfusionProbabilityRule struct {
	LM             LanguageModel
	Grams          int
	Exceptions     []string
	WordToPairs    map[string][]*rules.ConfusionPair
	DefaultOff     bool
	RuleIDOverride string
}

func NewConfusionProbabilityRule(lm LanguageModel, grams int) *ConfusionProbabilityRule {
	if grams <= 0 {
		grams = 3
	}
	return &ConfusionProbabilityRule{LM: lm, Grams: grams}
}

func (r *ConfusionProbabilityRule) GetID() string {
	if r != nil && r.RuleIDOverride != "" {
		return r.RuleIDOverride
	}
	return ConfusionRuleID
}

func (r *ConfusionProbabilityRule) SetWordToPairs(m map[string][]*rules.ConfusionPair) {
	r.WordToPairs = m
}
