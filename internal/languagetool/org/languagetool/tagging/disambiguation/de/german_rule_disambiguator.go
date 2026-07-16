package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// GermanRuleDisambiguator ports
// org.languagetool.tagging.disambiguation.rules.de.GermanRuleDisambiguator.
// Optional stages: MultitokenIgnore, Multitoken2, Rules (XML deferred).
type GermanRuleDisambiguator struct {
	disambiguation.AbstractDisambiguator
	MultitokenIgnore func(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	Multitoken2      func(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	Rules            func(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

func NewGermanRuleDisambiguator() *GermanRuleDisambiguator {
	return &GermanRuleDisambiguator{}
}

func (d *GermanRuleDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	s := input
	if d.MultitokenIgnore != nil {
		s = d.MultitokenIgnore(s)
	}
	if d.Multitoken2 != nil {
		s = d.Multitoken2(s)
	}
	if d.Rules != nil {
		s = d.Rules(s)
	}
	return s
}
