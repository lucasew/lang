package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

type sentenceStep interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// SpanishHybridDisambiguator ports
// org.languagetool.tagging.disambiguation.es.SpanishHybridDisambiguator:
// spelling_global → /es/multiwords.txt → XmlRuleDisambiguator(lang, true).
type SpanishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	GlobalChunker sentenceStep
	Chunker       sentenceStep
	Rules         sentenceStep
}

// NewSpanishHybridDisambiguator builds stages Java constructs as final fields.
// Chunker is wired from official /es/multiwords.txt when discoverable
// (Java MultiWordChunker.getInstance("/es/multiwords.txt", true, true, false)
// + setRemovePreviousTags(true); no setIgnoreSpelling).
// GlobalChunker and Rules are left nil here (separate sectors).
func NewSpanishHybridDisambiguator() *SpanishHybridDisambiguator {
	d := &SpanishHybridDisambiguator{}
	if mw := SpanishMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	return d
}

func (d *SpanishHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
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

var _ disambiguation.Disambiguator = (*SpanishHybridDisambiguator)(nil)
