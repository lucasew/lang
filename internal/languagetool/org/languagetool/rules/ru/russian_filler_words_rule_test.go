package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianFillerWordsRule(t *testing.T) {
	rule := NewRussianFillerWordsRule(nil)
	require.Equal(t, "FILLER_WORDS_RU", rule.GetID())
	matches := rule.Match(languagetool.AnalyzePlain("Ну, эээ, не знаю."))
	require.Equal(t, 1, len(matches))
}
