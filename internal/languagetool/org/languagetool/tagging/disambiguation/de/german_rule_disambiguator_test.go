package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

type stepFunc func(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence

func (f stepFunc) Disambiguate(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return f(s)
}

func TestGermanRuleDisambiguator(t *testing.T) {
	d := NewGermanRuleDisambiguator()
	s := languagetool.AnalyzePlain("Hallo Welt")
	require.Equal(t, s, d.Disambiguate(s))
	called := false
	d.Rules = stepFunc(func(in *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		called = true
		return in
	})
	d.Disambiguate(s)
	require.True(t, called)
}
