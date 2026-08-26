package it

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java ItalianSynthesizer resource paths (constructor super args; no SOR).
const (
	ResourceFilename = "/it/italian_synth.dict"
	TagsFileName     = "/it/italian_tags.txt"
)

// ItalianSynthesizer ports org.languagetool.synthesis.it.ItalianSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type ItalianSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewItalianSynthesizer ports private ItalianSynthesizer() / deprecated Language ctor.
// Java: super("/it/italian_synth.dict", "/it/italian_tags.txt", "it")
func NewItalianSynthesizer(manual *synthesis.ManualSynthesizer) *ItalianSynthesizer {
	base := synthesis.NewBaseSynthesizer("it", manual)
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	return &ItalianSynthesizer{BaseSynthesizer: base}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *ItalianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *ItalianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*ItalianSynthesizer)(nil)
