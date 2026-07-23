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

// NewGermanRuleDisambiguator builds the hybrid stages Java constructs as final fields.
// MultitokenIgnore is wired from official /de/multitoken-ignore.txt when discoverable
// (Java multitokenSpeller + setIgnoreSpelling(true)).
// MultitokenGlobal is wired from official /spelling_global.txt when discoverable
// (Java multitokenSpeller3 + setIgnoreSpelling(true); allowFirstCapitalized=false).
// Suggest/XML remain optional injectors until those loaders are wired the same way.
func NewGermanRuleDisambiguator() *GermanRuleDisambiguator {
	d := &GermanRuleDisambiguator{}
	if mw := GermanMultitokenIgnore(); mw != nil {
		d.MultitokenIgnore = mw
	}
	if mw := GermanMultitokenGlobal(); mw != nil {
		d.MultitokenGlobal = mw
	}
	return d
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
