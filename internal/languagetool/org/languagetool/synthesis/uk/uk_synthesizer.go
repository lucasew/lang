package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

type UkrainianSynthesizer struct{ *synthesis.BaseSynthesizer }

func NewUkrainianSynthesizer(m *synthesis.ManualSynthesizer) *UkrainianSynthesizer {
	b := synthesis.NewBaseSynthesizer("uk", m)
	// Java UkrainianSynthesizer: /uk/ukrainian_synth.dict + /uk/ukrainian_tags.txt
	b.ResourceFileName = "/uk/ukrainian_synth.dict"
	b.TagFileName = "/uk/ukrainian_tags.txt"
	return &UkrainianSynthesizer{BaseSynthesizer: b}
}
func (s *UkrainianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *UkrainianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*UkrainianSynthesizer)(nil)
