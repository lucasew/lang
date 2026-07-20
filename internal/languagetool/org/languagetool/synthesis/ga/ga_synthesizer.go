package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

type IrishSynthesizer struct { *synthesis.BaseSynthesizer }

func NewIrishSynthesizer(m *synthesis.ManualSynthesizer) *IrishSynthesizer {
	b := synthesis.NewBaseSynthesizer("ga", m)
	b.ResourceFileName = "/ga/irish_synth.dict"
	b.TagFileName = "/ga/irish_tags.txt"
	return &IrishSynthesizer{BaseSynthesizer: b}
}
func (s *IrishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *IrishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}
var _ synthesis.Synthesizer = (*IrishSynthesizer)(nil)
