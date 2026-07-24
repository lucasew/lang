package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

type SwedishSynthesizer struct { *synthesis.BaseSynthesizer }

func NewSwedishSynthesizer(m *synthesis.ManualSynthesizer) *SwedishSynthesizer {
	b := synthesis.NewBaseSynthesizer("sv", m)
	b.ResourceFileName = "/sv/swedish_synth.dict"
	b.TagFileName = "/sv/swedish_synth.dict_tags.txt"
	b.SorFileName = "/sv/sv.sor"
	return &SwedishSynthesizer{BaseSynthesizer: b}
}
func (s *SwedishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *SwedishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}
var _ synthesis.Synthesizer = (*SwedishSynthesizer)(nil)
