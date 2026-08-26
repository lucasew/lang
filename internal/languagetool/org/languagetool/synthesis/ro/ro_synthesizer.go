package ro

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java RomanianSynthesizer resource paths (constructor super args; no SOR).
const (
	ResourceFilename = "/ro/romanian_synth.dict"
	TagsFileName     = "/ro/romanian_tags.txt"
)

// RomanianSynthesizer ports org.languagetool.synthesis.ro.RomanianSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type RomanianSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewRomanianSynthesizer ports private RomanianSynthesizer() / deprecated Language ctor.
// Java: super("/ro/romanian_synth.dict", "/ro/romanian_tags.txt", "ro")
func NewRomanianSynthesizer(manual *synthesis.ManualSynthesizer) *RomanianSynthesizer {
	base := synthesis.NewBaseSynthesizer("ro", manual)
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	return &RomanianSynthesizer{BaseSynthesizer: base}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *RomanianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *RomanianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*RomanianSynthesizer)(nil)
