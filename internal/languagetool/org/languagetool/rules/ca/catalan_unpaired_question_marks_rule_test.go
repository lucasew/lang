package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanUnpairedQuestionMarksRule(t *testing.T) {
	rule := NewCatalanUnpairedQuestionMarksRule(nil)
	// Missing opening ¿
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Què passa?")})
	require.Equal(t, 1, len(matches))
	require.Equal(t, "¿Què", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("¿Què passa?")})))
}

func TestCatalanUnpairedExclamationMarksRule(t *testing.T) {
	rule := NewCatalanUnpairedExclamationMarksRule(nil)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Quina alegria!")})
	require.Equal(t, 1, len(matches))
	require.Equal(t, "¡Quina", matches[0].GetSuggestedReplacements()[0])
}
