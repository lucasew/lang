package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// CatalanSynthesizer ports synthesis.ca.CatalanSynthesizer.
type CatalanSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewCatalanSynthesizer(manual *synthesis.ManualSynthesizer) *CatalanSynthesizer {
	base := synthesis.NewBaseSynthesizer("ca", manual)
	base.ResourceFileName = "/ca/ca-ES_synth.dict"
	base.TagFileName = "/ca/ca-ES_tags.txt"
	base.SorFileName = "/ca/ca.sor"
	return &CatalanSynthesizer{BaseSynthesizer: base}
}

func (s *CatalanSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *CatalanSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*CatalanSynthesizer)(nil)
