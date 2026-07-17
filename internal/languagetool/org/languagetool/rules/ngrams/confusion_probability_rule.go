package ngrams

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ConfusionRuleID ports deprecated ConfusionProbabilityRule.RULE_ID.
const ConfusionRuleID = "CONFUSION_RULE"

// MinCoverage ports ConfusionProbabilityRule.MIN_COVERAGE.
const MinCoverage = 0.5

// ConfusionProbabilityRule is a metadata + data-holder port of
// org.languagetool.rules.ngrams.ConfusionProbabilityRule.
// Full n-gram scoring Match is deferred; exception/lookup helpers are green.
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

// IsLocalException reports whether text contains a configured exception phrase (soft).
func (r *ConfusionProbabilityRule) IsLocalException(text string) bool {
	if r == nil || text == "" {
		return false
	}
	low := strings.ToLower(text)
	for _, ex := range r.Exceptions {
		if ex != "" && strings.Contains(low, strings.ToLower(ex)) {
			return true
		}
	}
	return false
}

// PairsFor returns confusion pairs for a surface word (case-insensitive key).
func (r *ConfusionProbabilityRule) PairsFor(word string) []*rules.ConfusionPair {
	if r == nil || r.WordToPairs == nil {
		return nil
	}
	if p := r.WordToPairs[word]; len(p) > 0 {
		return p
	}
	return r.WordToPairs[strings.ToLower(word)]
}
