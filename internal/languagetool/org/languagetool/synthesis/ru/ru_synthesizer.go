package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java RussianSynthesizer.RESOURCE_FILENAME / TAGS_FILE_NAME.
const (
	ResourceFilename = "/ru/russian_synth.dict"
	TagsFileName     = "/ru/tags_russian.txt"
)

// RussianSynthesizer ports org.languagetool.synthesis.ru.RussianSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type RussianSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewRussianSynthesizer ports private RussianSynthesizer() / deprecated Language ctor.
// Java: super(RESOURCE_FILENAME, TAGS_FILE_NAME, "ru")
func NewRussianSynthesizer(m *synthesis.ManualSynthesizer) *RussianSynthesizer {
	b := synthesis.NewBaseSynthesizer("ru", m)
	b.ResourceFileName = ResourceFilename
	b.TagFileName = TagsFileName
	return &RussianSynthesizer{BaseSynthesizer: b}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *RussianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize RE path).
func (s *RussianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*RussianSynthesizer)(nil)
