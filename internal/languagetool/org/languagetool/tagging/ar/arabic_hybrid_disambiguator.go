package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// ArabicHybridDisambiguator ports org.languagetool.tagging.ar.ArabicHybridDisambiguator.
// Chains multiword chunker then optional secondary disambiguator.
type ArabicHybridDisambiguator struct {
	Chunker       disambiguation.Disambiguator
	Disambiguator disambiguation.Disambiguator
}

func NewArabicHybridDisambiguator(chunker, secondary disambiguation.Disambiguator) *ArabicHybridDisambiguator {
	return &ArabicHybridDisambiguator{Chunker: chunker, Disambiguator: secondary}
}

// NewDefaultArabicHybridDisambiguator uses empty multiword list (dict path deferred).
func NewDefaultArabicHybridDisambiguator() *ArabicHybridDisambiguator {
	return NewArabicHybridDisambiguator(
		disambiguation.NewMultiWordChunker(nil, disambiguation.MultiWordChunkerSettings{}),
		nil,
	)
}

func (d *ArabicHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if d == nil || input == nil {
		return input
	}
	out := input
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Disambiguator != nil {
		out = d.Disambiguator.Disambiguate(out)
	}
	return out
}

func (d *ArabicHybridDisambiguator) PreDisambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return input
}

var _ disambiguation.Disambiguator = (*ArabicHybridDisambiguator)(nil)
