package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// RussianSynthesizer ports synthesis.ru.RussianSynthesizer.
type RussianSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewRussianSynthesizer(manual *synthesis.ManualSynthesizer) *RussianSynthesizer {
	base := synthesis.NewBaseSynthesizer("ru", manual)
	base.ResourceFileName = "/ru/ru_synth.dict"
	base.TagFileName = "/ru/ru_tags.txt"
	return &RussianSynthesizer{BaseSynthesizer: base}
}

func (s *RussianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *RussianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*RussianSynthesizer)(nil)
