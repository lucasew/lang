package sk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

type SlovakHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Inner interface { Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence }
}

func NewSlovakHybridDisambiguator() *SlovakHybridDisambiguator { return &SlovakHybridDisambiguator{} }

func (d *SlovakHybridDisambiguator) Disambiguate(in *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if in == nil { return nil }
	if d.Inner != nil { return d.Inner.Disambiguate(in) }
	return in
}
var _ disambiguation.Disambiguator = (*SlovakHybridDisambiguator)(nil)
