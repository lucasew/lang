package it

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// ItalianSynthesizer ports synthesis.it.ItalianSynthesizer.
type ItalianSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewItalianSynthesizer(manual *synthesis.ManualSynthesizer) *ItalianSynthesizer {
	base := synthesis.NewBaseSynthesizer("it", manual)
	// Java ItalianSynthesizer: "/it/italian_synth.dict", "/it/italian_tags.txt"
	base.ResourceFileName = "/it/italian_synth.dict"
	base.TagFileName = "/it/italian_tags.txt"
	return &ItalianSynthesizer{BaseSynthesizer: base}
}

func (s *ItalianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *ItalianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*ItalianSynthesizer)(nil)
