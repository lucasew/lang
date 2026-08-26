package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java IrishSynthesizer resource paths (constructor super args; no SOR).
const (
	ResourceFilename = "/ga/irish_synth.dict"
	TagsFileName     = "/ga/irish_tags.txt"
)

// IrishSynthesizer ports org.languagetool.synthesis.ga.IrishSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type IrishSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewIrishSynthesizer ports private IrishSynthesizer() / deprecated Language ctor.
// Java: super("/ga/irish_synth.dict", "/ga/irish_tags.txt", "ga")
func NewIrishSynthesizer(manual *synthesis.ManualSynthesizer) *IrishSynthesizer {
	base := synthesis.NewBaseSynthesizer("ga", manual)
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	return &IrishSynthesizer{BaseSynthesizer: base}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *IrishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *IrishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*IrishSynthesizer)(nil)
