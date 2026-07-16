package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// PolishSynthesizer ports synthesis.pl.PolishSynthesizer.
type PolishSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewPolishSynthesizer(manual *synthesis.ManualSynthesizer) *PolishSynthesizer {
	base := synthesis.NewBaseSynthesizer("pl", manual)
	base.ResourceFileName = "/pl/pl_synth.dict"
	base.TagFileName = "/pl/pl_tags.txt"
	return &PolishSynthesizer{BaseSynthesizer: base}
}

func (s *PolishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *PolishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*PolishSynthesizer)(nil)
