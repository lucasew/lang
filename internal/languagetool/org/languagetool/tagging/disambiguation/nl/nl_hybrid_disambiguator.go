package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

type sentenceStep interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// DutchHybridDisambiguator ports org.languagetool.tagging.nl.DutchHybridDisambiguator
// (also mirrored under tagging/disambiguation/nl for package layout).
// Java: spelling_global → multiwords (tagForNotAddingTags) → XmlRuleDisambiguator.
type DutchHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	GlobalChunker sentenceStep
	Chunker       sentenceStep
	Rules         sentenceStep
}

func NewDutchHybridDisambiguator() *DutchHybridDisambiguator {
	return &DutchHybridDisambiguator{}
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
