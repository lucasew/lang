package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanConfusionProbabilityRule is a stand-in without n-gram language models.
// Match is a no-op until confusion pair LMs are ported.
type GermanConfusionProbabilityRule struct {
	Messages map[string]string
}

func NewGermanConfusionProbabilityRule(messages map[string]string) *GermanConfusionProbabilityRule {
	return &GermanConfusionProbabilityRule{Messages: messages}
}

func (r *GermanConfusionProbabilityRule) GetID() string {
	return "DE_CONFUSION_RULE"
}

func (r *GermanConfusionProbabilityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return nil
}
