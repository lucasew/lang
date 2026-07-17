package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// UkrainianHybridDisambiguator ports tagging.disambiguation.uk.UkrainianHybridDisambiguator.
type UkrainianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker runs first (multiword); Inner/Disambiguator second.
	Chunker disambiguation.Disambiguator
	Inner   disambiguation.Disambiguator
}

func NewUkrainianHybridDisambiguator() *UkrainianHybridDisambiguator {
	return &UkrainianHybridDisambiguator{
		Chunker: NewUkrainianMultiwordChunker(nil),
	}
}

// NewUkrainianHybridDisambiguatorWith sets optional stages.
func NewUkrainianHybridDisambiguatorWith(chunker, secondary disambiguation.Disambiguator) *UkrainianHybridDisambiguator {
	return &UkrainianHybridDisambiguator{Chunker: chunker, Inner: secondary}
}

func (d *UkrainianHybridDisambiguator) Disambiguate(in *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if in == nil {
		return nil
	}
	out := in
	if d != nil && d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d != nil && d.Inner != nil {
		out = d.Inner.Disambiguate(out)
	}
	// soft: strip v_mis when other cases remain (prep context deferred)
	if out != nil {
		for _, tok := range out.GetTokensWithoutWhitespace() {
			RemoveVmisReadings(tok)
		}
	}
	return out
}

var _ disambiguation.Disambiguator = (*UkrainianHybridDisambiguator)(nil)
