package it

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// ItalianRuleDisambiguator ports
// org.languagetool.tagging.disambiguation.rules.it.ItalianRuleDisambiguator.
type ItalianRuleDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Rules optional XML rule disambiguator.
	Rules func(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

func NewItalianRuleDisambiguator() *ItalianRuleDisambiguator {
	return &ItalianRuleDisambiguator{}
}

func (d *ItalianRuleDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	if d.Rules != nil {
		return d.Rules(input)
	}
	return input
}
