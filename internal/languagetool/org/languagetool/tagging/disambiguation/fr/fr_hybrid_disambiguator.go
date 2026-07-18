package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// sentenceStep is a Disambiguate-capable stage (MultiWordChunker / XmlRuleDisambiguator).
type sentenceStep interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// FrenchHybridDisambiguator ports
// org.languagetool.tagging.disambiguation.fr.FrenchHybridDisambiguator:
// spelling_global → /fr/multiwords.txt → XmlRuleDisambiguator(lang, true).
type FrenchHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// GlobalChunker is Java chunkerGlobal (spelling_global.txt).
	GlobalChunker sentenceStep
	// Chunker is Java /fr/multiwords.txt MultiWordChunker.
	Chunker sentenceStep
	// Rules is Java XmlRuleDisambiguator.
	Rules sentenceStep
}

func NewFrenchHybridDisambiguator() *FrenchHybridDisambiguator {
	return &FrenchHybridDisambiguator{}
}

func (d *FrenchHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	out := input
	// Java: disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(...)))
	if d.GlobalChunker != nil {
		out = d.GlobalChunker.Disambiguate(out)
	}
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*FrenchHybridDisambiguator)(nil)
