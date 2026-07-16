package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// DutchHybridDisambiguator ports org.languagetool.tagging.nl.DutchHybridDisambiguator.
// Full MultiWordChunker + XML rule chain is deferred; optional stages are pluggable.
type DutchHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker optional multi-word stage.
	Chunker func(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	// Rules optional XML/rule disambiguation.
	Rules func(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

func NewDutchHybridDisambiguator() *DutchHybridDisambiguator {
	return &DutchHybridDisambiguator{}
}

func (d *DutchHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	s := input
	if d.Chunker != nil {
		s = d.Chunker(s)
	}
	if d.Rules != nil {
		s = d.Rules(s)
	}
	return s
}
