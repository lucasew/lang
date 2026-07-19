package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// UkrainianHybridDisambiguator ports tagging.disambiguation.uk.UkrainianHybridDisambiguator.
type UkrainianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker runs first (multiword); Inner/Disambiguator second (XML rules).
	Chunker disambiguation.Disambiguator
	Inner   disambiguation.Disambiguator
	// Simple is preDisambiguate (rare-form + dups maps).
	Simple *SimpleDisambiguator
}

func NewUkrainianHybridDisambiguator() *UkrainianHybridDisambiguator {
	return &UkrainianHybridDisambiguator{
		Chunker: NewUkrainianMultiwordChunker(nil),
		Simple:  NewSimpleDisambiguator(),
	}
}

// NewUkrainianHybridDisambiguatorWith sets optional stages (simple left nil).
func NewUkrainianHybridDisambiguatorWith(chunker, secondary disambiguation.Disambiguator) *UkrainianHybridDisambiguator {
	return &UkrainianHybridDisambiguator{Chunker: chunker, Inner: secondary, Simple: NewSimpleDisambiguator()}
}

func (d *UkrainianHybridDisambiguator) Disambiguate(in *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if in == nil {
		return nil
	}
	out := in
	// Java preDisambiguate: simpleDisambiguator.removeRareForms (+ more hybrid filters later)
	if d != nil && d.Simple != nil {
		out = d.Simple.Disambiguate(out)
	}
	if d != nil && d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d != nil && d.Inner != nil {
		out = d.Inner.Disambiguate(out)
	}
	// soft context rules (full XML disambig still optional via Inner)
	if out != nil {
		RetagInitials(out)
		DisambiguateSt(out)
		DisambiguatePronPos(out)
		RetagFemNames(out)
		RemoveInanimVKly(out)
		RemoveLowerCaseHomonymsForAbbreviations(out)
		RemoveLowerCaseBadForUpperCaseGood(out)
		RemovePluralForNames(out)
		RemoveVerbImpr(out)
		PreferVocativeWhenBang(out)
		for _, tok := range out.GetTokensWithoutWhitespace() {
			RemoveVmisReadings(tok)
		}
	}
	return out
}

var _ disambiguation.Disambiguator = (*UkrainianHybridDisambiguator)(nil)
