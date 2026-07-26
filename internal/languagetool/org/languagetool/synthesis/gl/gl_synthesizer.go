package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java GalicianSynthesizer resource paths (constructor super args; no SOR).
const (
	ResourceFilename = "/gl/galician_synth.dict"
	TagsFileName     = "/gl/galician_tags.txt"
)

// GalicianSynthesizer ports org.languagetool.synthesis.gl.GalicianSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type GalicianSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewGalicianSynthesizer ports private GalicianSynthesizer() / deprecated Language ctor.
// Java: super("/gl/galician_synth.dict", "/gl/galician_tags.txt", "gl")
func NewGalicianSynthesizer(manual *synthesis.ManualSynthesizer) *GalicianSynthesizer {
	base := synthesis.NewBaseSynthesizer("gl", manual)
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	return &GalicianSynthesizer{BaseSynthesizer: base}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *GalicianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *GalicianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*GalicianSynthesizer)(nil)
