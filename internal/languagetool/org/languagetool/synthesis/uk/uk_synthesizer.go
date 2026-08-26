package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java UkrainianSynthesizer.RESOURCE_FILENAME / TAGS_FILE_NAME.
const (
	ResourceFilename = "/uk/ukrainian_synth.dict"
	TagsFileName     = "/uk/ukrainian_tags.txt"
)

// UkrainianSynthesizer ports org.languagetool.synthesis.uk.UkrainianSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type UkrainianSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewUkrainianSynthesizer ports private UkrainianSynthesizer() / deprecated Language ctor.
func NewUkrainianSynthesizer(m *synthesis.ManualSynthesizer) *UkrainianSynthesizer {
	b := synthesis.NewBaseSynthesizer("uk", m)
	// Java: super(RESOURCE_FILENAME, TAGS_FILE_NAME, "uk")
	b.ResourceFileName = ResourceFilename
	b.TagFileName = TagsFileName
	return &UkrainianSynthesizer{BaseSynthesizer: b}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *UkrainianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *UkrainianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*UkrainianSynthesizer)(nil)
