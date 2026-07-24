package da

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

type DanishSynthesizer struct { *synthesis.BaseSynthesizer }

func NewDanishSynthesizer(m *synthesis.ManualSynthesizer) *DanishSynthesizer {
	b := synthesis.NewBaseSynthesizer("da", m)
	b.ResourceFileName = "/da/da_synth.dict"
	b.TagFileName = "/da/da_tags.txt"
	return &DanishSynthesizer{BaseSynthesizer: b}
}
func (s *DanishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *DanishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}
var _ synthesis.Synthesizer = (*DanishSynthesizer)(nil)
