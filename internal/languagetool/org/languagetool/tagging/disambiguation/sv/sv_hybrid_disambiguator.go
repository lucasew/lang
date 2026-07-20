package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// SwedishHybridDisambiguator ports org.languagetool.tagging.disambiguation.sv.SwedishHybridDisambiguator.
// Java: chunker.disambiguate(disambiguator.disambiguate(input)) — XML first, then multiwords.
type SwedishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

func NewSwedishHybridDisambiguator() *SwedishHybridDisambiguator {
	return &SwedishHybridDisambiguator{}
}

func (d *SwedishHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	out := input
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*SwedishHybridDisambiguator)(nil)
