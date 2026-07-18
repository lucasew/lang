package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// EnglishHybridDisambiguator ports
// org.languagetool.tagging.en.EnglishHybridDisambiguator:
// spelling_global MultiWordChunker, then /en/multiwords.txt chunker, then XML rules.
type EnglishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// GlobalChunker optional spelling_global.txt chunker (Java chunkerGlobal first).
	GlobalChunker interface {
		Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Chunker optional multi-word chunker applied after GlobalChunker.
	Chunker interface {
		Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// RulesDisambiguator optional XML rule disambiguator.
	RulesDisambiguator interface {
		Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

func NewEnglishHybridDisambiguator() *EnglishHybridDisambiguator {
	return &EnglishHybridDisambiguator{}
}

func (d *EnglishHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
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
	if d.RulesDisambiguator != nil {
		out = d.RulesDisambiguator.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*EnglishHybridDisambiguator)(nil)
