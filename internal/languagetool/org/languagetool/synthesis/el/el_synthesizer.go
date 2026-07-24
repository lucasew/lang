package el

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

type GreekSynthesizer struct { *synthesis.BaseSynthesizer }

func NewGreekSynthesizer(m *synthesis.ManualSynthesizer) *GreekSynthesizer {
	b := synthesis.NewBaseSynthesizer("el", m)
	b.ResourceFileName = "/el/greek_synth.dict"
	b.TagFileName = "/el/greek_tags.txt"
	return &GreekSynthesizer{BaseSynthesizer: b}
}
func (s *GreekSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *GreekSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}
var _ synthesis.Synthesizer = (*GreekSynthesizer)(nil)
