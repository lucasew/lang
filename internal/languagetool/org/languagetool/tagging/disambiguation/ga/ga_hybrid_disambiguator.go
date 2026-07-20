package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// IrishHybridDisambiguator ports org.languagetool.tagging.disambiguation.ga.IrishHybridDisambiguator.
// Java: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords first, then XML.
type IrishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

func NewIrishHybridDisambiguator() *IrishHybridDisambiguator {
	return &IrishHybridDisambiguator{}
}

func (d *IrishHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	out := input
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*IrishHybridDisambiguator)(nil)
