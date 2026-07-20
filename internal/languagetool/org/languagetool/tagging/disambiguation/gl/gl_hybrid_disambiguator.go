package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// GalicianHybridDisambiguator ports org.languagetool.tagging.disambiguation.gl.GalicianHybridDisambiguator.
// Java: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords first, then XML.
type GalicianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

func NewGalicianHybridDisambiguator() *GalicianHybridDisambiguator {
	return &GalicianHybridDisambiguator{}
}

func (d *GalicianHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
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

var _ disambiguation.Disambiguator = (*GalicianHybridDisambiguator)(nil)
