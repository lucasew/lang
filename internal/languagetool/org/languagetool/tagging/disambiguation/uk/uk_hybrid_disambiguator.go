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
	// Java: multiwords + XmlRuleDisambiguator(Ukrainian) + SimpleDisambiguator pre
	return &UkrainianHybridDisambiguator{
		Chunker: UkrainianMultiwordChunkerDefault(),
		Inner:   LoadDefaultUkrainianXmlDisambiguator(),
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
	// Java preDisambiguate BEFORE chunker/XML (UkrainianHybridDisambiguator.preDisambiguate).
	if d != nil && d.Simple != nil {
		out = d.Simple.Disambiguate(out)
	}
	if out != nil {
		DisambiguateYih(out)
		// Java removeVmis is sentence-level (startCheck + V_MIS_PREPS early return)
		RemoveVmis(out)
		RetagFemNames(out)
		RetagInitials(out)
		RetagUnknownInitials(out)
		RemoveInanimVKly(out)
		RemovePluralForNames(out)
		RemoveLowerCaseHomonymsForAbbreviations(out)
		RemoveLowerCaseBadForUpperCaseGood(out)
		DisambiguateSt(out)
		DisambiguatePronPos(out)
		RetagPluralProp(out)
		RemoveVerbImpr(out)
		// PreferVocativeWhenBang is not in Java preDisambiguate list (helper remains available).
	}
	// Java: disambiguator.disambiguate(chunker.disambiguate(input))
	if d != nil && d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d != nil && d.Inner != nil {
		out = d.Inner.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*UkrainianHybridDisambiguator)(nil)
