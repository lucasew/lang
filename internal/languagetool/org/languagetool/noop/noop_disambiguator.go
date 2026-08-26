package noop

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// NoopDisambiguator ports org.languagetool.noop.NoopDisambiguator.
type NoopDisambiguator struct {
	disambiguation.AbstractDisambiguator
}

func NewNoopDisambiguator() *NoopDisambiguator {
	return &NoopDisambiguator{}
}

func (NoopDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return input
}
