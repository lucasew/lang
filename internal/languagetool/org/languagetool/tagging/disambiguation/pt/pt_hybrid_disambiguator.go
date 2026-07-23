package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

type sentenceStep interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// PortugueseHybridDisambiguator ports
// org.languagetool.tagging.disambiguation.pt.PortugueseHybridDisambiguator:
// spelling_global → /pt/multiwords.txt → XmlRuleDisambiguator(lang, true).
type PortugueseHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	GlobalChunker sentenceStep
	Chunker       sentenceStep
	Rules         sentenceStep
}

// NewPortugueseHybridDisambiguator builds stages Java constructs as final fields.
// Chunker is wired from official /pt/multiwords.txt when discoverable
// (getInstance("/pt/multiwords.txt", true, true, true)
// + setRemovePreviousTags(true) + setIgnoreSpelling(true)).
// GlobalChunker and Rules stay nil until their sectors land.
func NewPortugueseHybridDisambiguator() *PortugueseHybridDisambiguator {
	d := &PortugueseHybridDisambiguator{}
	if mw := PortugueseMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	return d
}

func (d *PortugueseHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
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

var _ disambiguation.Disambiguator = (*PortugueseHybridDisambiguator)(nil)
