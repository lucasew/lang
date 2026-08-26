package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java SwedishSynthesizer resource paths (constructor super args).
const (
	ResourceFilename = "/sv/swedish_synth.dict"
	TagsFileName     = "/sv/swedish_synth.dict_tags.txt"
	SorFileName      = "/sv/sv.sor"
)

// SwedishSynthesizer ports org.languagetool.synthesis.sv.SwedishSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type SwedishSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewSwedishSynthesizer ports private SwedishSynthesizer() / deprecated Language ctor.
// Java: super(SOR_FILE_NAME, RESOURCE_FILENAME, TAGS_FILE_NAME, "sv")
// SOR = "/sv/sv.sor", RESOURCE = "/sv/swedish_synth.dict", TAGS = "/sv/swedish_synth.dict_tags.txt"
func NewSwedishSynthesizer(manual *synthesis.ManualSynthesizer) *SwedishSynthesizer {
	base := synthesis.NewBaseSynthesizer("sv", manual)
	base.SorFileName = SorFileName
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	return &SwedishSynthesizer{BaseSynthesizer: base}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *SwedishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *SwedishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*SwedishSynthesizer)(nil)
