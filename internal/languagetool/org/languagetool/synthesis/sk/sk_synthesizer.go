package sk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java SlovakSynthesizer resource paths (constructor super args; no SOR).
const (
	ResourceFilename = "/sk/slovak_synth.dict"
	TagsFileName     = "/sk/slovak_tags.txt"
)

// SlovakSynthesizer ports org.languagetool.synthesis.sk.SlovakSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type SlovakSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewSlovakSynthesizer ports private SlovakSynthesizer() / deprecated Language ctor.
// Java: super("/sk/slovak_synth.dict", "/sk/slovak_tags.txt", "sk")
func NewSlovakSynthesizer(manual *synthesis.ManualSynthesizer) *SlovakSynthesizer {
	base := synthesis.NewBaseSynthesizer("sk", manual)
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	return &SlovakSynthesizer{BaseSynthesizer: base}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *SlovakSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *SlovakSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*SlovakSynthesizer)(nil)
