package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanDoublePunctuationRule ports org.languagetool.rules.de.GermanDoublePunctuationRule.
type GermanDoublePunctuationRule struct {
	*rules.DoublePunctuationRule
}

func NewGermanDoublePunctuationRule(messages map[string]string) *GermanDoublePunctuationRule {
	base := rules.NewDoublePunctuationRule(messages)
	base.RuleID = "DE_DOUBLE_PUNCTUATION"
	base.DotMessage = "Zwei aufeinander folgende Punkte. Auch wenn ein Satz mit einer Abkürzung endet, " +
		"endet er nur mit einem Punkt (§103 Regelwerk)."
	return &GermanDoublePunctuationRule{DoublePunctuationRule: base}
}

func (r *GermanDoublePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.DoublePunctuationRule.Match(sentence)
}
