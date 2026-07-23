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
// GlobalChunker is wired from official /spelling_global.txt when discoverable
// (getInstance("/spelling_global.txt", false, true, false, "NPCN000");
// no setIgnoreSpelling; no setRemovePreviousTags).
// Chunker is wired from official /ca/multiwords.txt when discoverable
// (getInstance("/ca/multiwords.txt", true, true, false)
// + setRemovePreviousTags(true); no setIgnoreSpelling).
// Rules is wired from XmlRuleDisambiguator(lang, true) when official ca + global
// XML load (useGlobalDisambiguation=true).
// Multitoken is always constructed (Java field initializer
// new CatalanMultitokenDisambiguator()); with nil IsMisspelled it no-ops like
// Java when speller == null — no invent dictionary.
func NewCatalanHybridDisambiguator() *CatalanHybridDisambiguator {
	d := &CatalanHybridDisambiguator{
		Multitoken: NewCatalanMultitokenDisambiguator(),
	}
	if g := CatalanGlobalChunker(); g != nil {
		d.GlobalChunker = g
	}
	if mw := CatalanMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	if r := CatalanXmlRuleDisambiguator(); r != nil {
		d.Rules = r
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
