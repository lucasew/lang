package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// FrenchSynthesizer ports synthesis.FrenchSynthesizer.
type FrenchSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewFrenchSynthesizer(manual *synthesis.ManualSynthesizer) *FrenchSynthesizer {
	base := synthesis.NewBaseSynthesizer("fr", manual)
	base.ResourceFileName = "/fr/french_synth.dict"
	base.TagFileName = "/fr/french_tags.txt"
	return &FrenchSynthesizer{BaseSynthesizer: base}
}

func (s *FrenchSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *FrenchSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*FrenchSynthesizer)(nil)
