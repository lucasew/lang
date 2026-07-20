package sk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

type SlovakSynthesizer struct { *synthesis.BaseSynthesizer }

func NewSlovakSynthesizer(m *synthesis.ManualSynthesizer) *SlovakSynthesizer {
	b := synthesis.NewBaseSynthesizer("sk", m)
	b.ResourceFileName = "/sk/slovak_synth.dict"
	b.TagFileName = "/sk/slovak_tags.txt"
	return &SlovakSynthesizer{BaseSynthesizer: b}
}
func (s *SlovakSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *SlovakSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}
var _ synthesis.Synthesizer = (*SlovakSynthesizer)(nil)
