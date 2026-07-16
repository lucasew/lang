package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

type IrishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Inner interface { Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence }
}

func NewIrishHybridDisambiguator() *IrishHybridDisambiguator { return &IrishHybridDisambiguator{} }

func (d *IrishHybridDisambiguator) Disambiguate(in *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if in == nil { return nil }
	if d.Inner != nil { return d.Inner.Disambiguate(in) }
	return in
}
var _ disambiguation.Disambiguator = (*IrishHybridDisambiguator)(nil)
