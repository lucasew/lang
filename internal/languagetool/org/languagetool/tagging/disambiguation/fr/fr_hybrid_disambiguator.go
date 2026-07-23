package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// sentenceStep is a Disambiguate-capable stage (MultiWordChunker / XmlRuleDisambiguator).
type sentenceStep interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// FrenchHybridDisambiguator ports
// org.languagetool.tagging.disambiguation.fr.FrenchHybridDisambiguator:
// spelling_global → /fr/multiwords.txt → XmlRuleDisambiguator(lang, true).
type FrenchHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// GlobalChunker is Java chunkerGlobal (spelling_global.txt).
	GlobalChunker sentenceStep
	// Chunker is Java /fr/multiwords.txt MultiWordChunker.
	Chunker sentenceStep
	// Rules is Java XmlRuleDisambiguator.
	Rules sentenceStep
}

// NewFrenchHybridDisambiguator builds stages Java constructs as final fields.
// GlobalChunker is wired from official /spelling_global.txt when discoverable
// (getInstance("/spelling_global.txt", false, true, false, tagForNotAddingTags)
// + setIgnoreSpelling(true)).
// Chunker is wired from official /fr/multiwords.txt when discoverable
// (getInstance("/fr/multiwords.txt", true, true, false)
// + setRemovePreviousTags(true); no setIgnoreSpelling).
// Rules remain nil until the XmlRuleDisambiguator sector lands.
func NewFrenchHybridDisambiguator() *FrenchHybridDisambiguator {
	d := &FrenchHybridDisambiguator{}
	if g := FrenchGlobalChunker(); g != nil {
		d.GlobalChunker = g
	}
	if mw := FrenchMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	return d
}

func (d *FrenchHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	out := input
	// Java: disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(...)))
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

var _ disambiguation.Disambiguator = (*FrenchHybridDisambiguator)(nil)
