package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// EnglishHybridDisambiguator ports
// org.languagetool.tagging.en.EnglishHybridDisambiguator:
// MultiWordChunker (optional) then XmlRuleDisambiguator (pluggable).
type EnglishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker optional multi-word chunker applied first.
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
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.RulesDisambiguator != nil {
		out = d.RulesDisambiguator.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*EnglishHybridDisambiguator)(nil)
