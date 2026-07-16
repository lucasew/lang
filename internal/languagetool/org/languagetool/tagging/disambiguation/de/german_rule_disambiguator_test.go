package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanRuleDisambiguator(t *testing.T) {
	d := NewGermanRuleDisambiguator()
	s := languagetool.AnalyzePlain("Hallo Welt")
	require.Equal(t, s, d.Disambiguate(s))
	called := false
	d.Rules = func(in *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
		called = true
		return in
	}
	d.Disambiguate(s)
	require.True(t, called)
}
