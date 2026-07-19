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

func (r *GermanDoublePunctuationRule) GetID() string {
	if r != nil && r.DoublePunctuationRule != nil {
		return r.DoublePunctuationRule.GetID()
	}
	return "DE_DOUBLE_PUNCTUATION"
}

// GetURL ports GermanDoublePunctuationRule constructor setUrl.
func (r *GermanDoublePunctuationRule) GetURL() string {
	return "https://dict.leo.org/grammatik/deutsch/Rechtschreibung/Amtlich/Interpunktion/pgf101-105.html#grammarpgf103"
}

func (r *GermanDoublePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	// Java attaches this (DE rule) so setUrl is visible on matches.
	ms := r.DoublePunctuationRule.Match(sentence)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
