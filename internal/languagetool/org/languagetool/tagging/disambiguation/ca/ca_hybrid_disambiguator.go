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

// NewCatalanHybridDisambiguator builds stages Java constructs as final fields.
// Chunker is wired from official /ca/multiwords.txt when discoverable
// (getInstance("/ca/multiwords.txt", true, true, false)
// + setRemovePreviousTags(true); no setIgnoreSpelling).
// GlobalChunker / Rules / Multitoken remain unwired until their sectors land.
func NewCatalanHybridDisambiguator() *CatalanHybridDisambiguator {
	d := &CatalanHybridDisambiguator{}
	if mw := CatalanMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	return d
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
