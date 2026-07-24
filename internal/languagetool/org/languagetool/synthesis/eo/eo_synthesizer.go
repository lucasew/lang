package eo

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

type EsperantoSynthesizer struct { *synthesis.BaseSynthesizer }

func NewEsperantoSynthesizer(m *synthesis.ManualSynthesizer) *EsperantoSynthesizer {
	b := synthesis.NewBaseSynthesizer("eo", m)
	b.ResourceFileName = "/eo/eo_synth.dict"
	b.TagFileName = "/eo/eo_tags.txt"
	return &EsperantoSynthesizer{BaseSynthesizer: b}
}
func (s *EsperantoSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *EsperantoSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}
var _ synthesis.Synthesizer = (*EsperantoSynthesizer)(nil)
