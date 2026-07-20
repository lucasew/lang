package sr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// SerbianHybridDisambiguator ports
// org.languagetool.tagging.disambiguation.sr.SerbianHybridDisambiguator:
// MultiWordChunker("/sr/multiwords.txt") defaults, then XmlRuleDisambiguator(Serbian) no global.
// Java order: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords then XML.
type SerbianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

func NewSerbianHybridDisambiguator() *SerbianHybridDisambiguator {
	return &SerbianHybridDisambiguator{}
}

func (d *SerbianHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	out := input
	// multiwords first, then XML
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*SerbianHybridDisambiguator)(nil)
