package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// sentenceStep is a Disambiguate-capable stage (MultiWordChunker / XmlRuleDisambiguator).
type sentenceStep interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// DutchHybridDisambiguator ports org.languagetool.tagging.nl.DutchHybridDisambiguator.
// Java: spelling_global → multiwords (tagForNotAddingTags) → XmlRuleDisambiguator.
type DutchHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	GlobalChunker sentenceStep
	Chunker       sentenceStep
	Rules         sentenceStep
}

// NewDutchHybridDisambiguator builds stages Java constructs as final fields.
// Chunker is wired from official /nl/multiwords.txt when discoverable
// (Java MultiWordChunker.getInstance(..., true, true, false, tagForNotAddingTags)
// + setIgnoreSpelling(true)). GlobalChunker and Rules remain optional injectors
// until those sectors land.
func NewDutchHybridDisambiguator() *DutchHybridDisambiguator {
	d := &DutchHybridDisambiguator{}
	if mw := DutchMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	return d
}

func (d *DutchHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	out := input
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

var _ disambiguation.Disambiguator = (*DutchHybridDisambiguator)(nil)
