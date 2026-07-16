package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestIsDigitHyphenAdj(t *testing.T) {
	require.True(t, IsDigitHyphenAdj("3-adische"))
	require.True(t, IsDigitHyphenAdj("2-fach"))
	require.False(t, IsDigitHyphenAdj("Kelassurier"))
	require.False(t, IsDigitHyphenAdj("adische"))
}

func TestMarkIgnoreSpellingPatterns(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("3-adische", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("System", nil, nil)),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	out := MarkIgnoreSpellingPatterns(sent)
	require.True(t, out.GetTokens()[0].IsIgnoredBySpeller())
	require.False(t, out.GetTokens()[2].IsIgnoredBySpeller())
}

func TestGermanRuleDisambiguator_WithIgnoreStage(t *testing.T) {
	// use disambiguation/de package via import path in GermanDisambiguationTest
}
