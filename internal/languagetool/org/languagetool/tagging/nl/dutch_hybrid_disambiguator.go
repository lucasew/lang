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
// GlobalChunker is wired from official /spelling_global.txt when discoverable
// (Java MultiWordChunker.getInstance(..., false, true, false, tagForNotAddingTags)
// + setIgnoreSpelling(true)). Chunker is wired from official /nl/multiwords.txt
// when discoverable (allowFirstCapitalized=true). Rules remain optional injectors
// until that sector lands.
func NewDutchHybridDisambiguator() *DutchHybridDisambiguator {
	d := &DutchHybridDisambiguator{}
	if g := DutchGlobalChunker(); g != nil {
		d.GlobalChunker = g
	}
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
