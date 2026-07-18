package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

type sentenceStep interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// GermanRuleDisambiguator ports
// org.languagetool.tagging.disambiguation.rules.de.GermanRuleDisambiguator.
// Java order (disambiguate with callback):
//
//	multitoken-ignore → spelling_global → multitoken-suggest → XmlRuleDisambiguator
type GermanRuleDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// MultitokenIgnore is Java multitokenSpeller (/de/multitoken-ignore.txt).
	MultitokenIgnore sentenceStep
	// MultitokenGlobal is Java multitokenSpeller3 (/spelling_global.txt).
	MultitokenGlobal sentenceStep
	// MultitokenSuggest is Java multitokenSpeller2 (/de/multitoken-suggest.txt).
	MultitokenSuggest sentenceStep
	// Rules is Java XmlRuleDisambiguator(lang, true).
	Rules sentenceStep
}

func NewGermanRuleDisambiguator() *GermanRuleDisambiguator {
	return &GermanRuleDisambiguator{}
}

func (d *GermanRuleDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	s := input
	// Java: disambiguator(multitokenSpeller2(multitokenSpeller3(multitokenSpeller(input))))
	if d.MultitokenIgnore != nil {
		s = d.MultitokenIgnore.Disambiguate(s)
	}
	if d.MultitokenGlobal != nil {
		s = d.MultitokenGlobal.Disambiguate(s)
	}
	if d.MultitokenSuggest != nil {
		s = d.MultitokenSuggest.Disambiguate(s)
	}
	if d.Rules != nil {
		s = d.Rules.Disambiguate(s)
	}
	return s
}

var _ disambiguation.Disambiguator = (*GermanRuleDisambiguator)(nil)
