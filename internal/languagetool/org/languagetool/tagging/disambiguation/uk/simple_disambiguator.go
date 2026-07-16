package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// SimpleDisambiguator ports a no-op / identity UK simple disambiguator surface.
type SimpleDisambiguator struct {
	disambiguation.AbstractDisambiguator
}

func NewSimpleDisambiguator() *SimpleDisambiguator { return &SimpleDisambiguator{} }

func (d *SimpleDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return input
}

var _ disambiguation.Disambiguator = (*SimpleDisambiguator)(nil)
