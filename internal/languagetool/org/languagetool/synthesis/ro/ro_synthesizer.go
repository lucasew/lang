package ro

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

type RomanianSynthesizer struct { *synthesis.BaseSynthesizer }

func NewRomanianSynthesizer(m *synthesis.ManualSynthesizer) *RomanianSynthesizer {
	b := synthesis.NewBaseSynthesizer("ro", m)
	b.ResourceFileName = "/ro/romanian_synth.dict"
	b.TagFileName = "/ro/romanian_tags.txt"
	return &RomanianSynthesizer{BaseSynthesizer: b}
}
func (s *RomanianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *RomanianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}
var _ synthesis.Synthesizer = (*RomanianSynthesizer)(nil)
