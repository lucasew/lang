package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

type GalicianSynthesizer struct { *synthesis.BaseSynthesizer }

func NewGalicianSynthesizer(m *synthesis.ManualSynthesizer) *GalicianSynthesizer {
	b := synthesis.NewBaseSynthesizer("gl", m)
	b.ResourceFileName = "/gl/galician_synth.dict"
	b.TagFileName = "/gl/galician_tags.txt"
	return &GalicianSynthesizer{BaseSynthesizer: b}
}
func (s *GalicianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *GalicianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}
var _ synthesis.Synthesizer = (*GalicianSynthesizer)(nil)
