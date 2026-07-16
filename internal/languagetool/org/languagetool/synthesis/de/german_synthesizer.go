package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

const (
	GermanSynthResource = "/de/german_synth.dict"
	GermanTagsFile      = "/de/german_tags.txt"
)

// GermanSynthesizer ports org.languagetool.synthesis.GermanSynthesizer.
type GermanSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewGermanSynthesizer(manual *synthesis.ManualSynthesizer) *GermanSynthesizer {
	base := synthesis.NewBaseSynthesizer("de", manual)
	base.ResourceFileName = GermanSynthResource
	base.TagFileName = GermanTagsFile
	return &GermanSynthesizer{BaseSynthesizer: base}
}

// Synthesize implements synthesis.Synthesizer.
func (s *GermanSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

func (s *GermanSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*GermanSynthesizer)(nil)
