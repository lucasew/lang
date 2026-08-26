package el

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java GreekSynthesizer resource paths (constructor super args; no SOR).
const (
	ResourceFilename = "/el/greek_synth.dict"
	TagsFileName     = "/el/greek_tags.txt"
)

// GreekSynthesizer ports org.languagetool.synthesis.el.GreekSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type GreekSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewGreekSynthesizer ports private GreekSynthesizer() / deprecated Language ctor.
// Java: super("/el/greek_synth.dict", "/el/greek_tags.txt", "el")
func NewGreekSynthesizer(manual *synthesis.ManualSynthesizer) *GreekSynthesizer {
	base := synthesis.NewBaseSynthesizer("el", manual)
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	return &GreekSynthesizer{BaseSynthesizer: base}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *GreekSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *GreekSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*GreekSynthesizer)(nil)
