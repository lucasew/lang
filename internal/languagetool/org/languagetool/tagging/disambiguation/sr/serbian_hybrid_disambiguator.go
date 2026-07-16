package sr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// SerbianHybridDisambiguator ports hybrid disambiguation for Serbian.
type SerbianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker func(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	Rules   func(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

func NewSerbianHybridDisambiguator() *SerbianHybridDisambiguator {
	return &SerbianHybridDisambiguator{}
}

func (d *SerbianHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	s := input
	if d.Chunker != nil {
		s = d.Chunker(s)
	}
	if d.Rules != nil {
		s = d.Rules(s)
	}
	return s
}
