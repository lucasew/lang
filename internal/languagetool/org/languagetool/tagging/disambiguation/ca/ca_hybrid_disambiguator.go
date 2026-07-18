package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

type sentenceStep interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// CatalanHybridDisambiguator ports
// org.languagetool.tagging.disambiguation.ca.CatalanHybridDisambiguator:
// spelling_global → multiwords → XmlRuleDisambiguator → CatalanMultitokenDisambiguator.
type CatalanHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	GlobalChunker sentenceStep
	Chunker       sentenceStep
	Rules         sentenceStep
	// Multitoken is Java CatalanMultitokenDisambiguator (after XML).
	Multitoken sentenceStep
}

func NewCatalanHybridDisambiguator() *CatalanHybridDisambiguator {
	return &CatalanHybridDisambiguator{}
}

func (d *CatalanHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	out := input
	// Java: multitokenDisambiguator(disambiguator(chunker(chunkerGlobal(input))))
	if d.GlobalChunker != nil {
		out = d.GlobalChunker.Disambiguate(out)
	}
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	if d.Multitoken != nil {
		out = d.Multitoken.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*CatalanHybridDisambiguator)(nil)
