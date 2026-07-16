package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianWordRootRepeatRule(t *testing.T) {
	rule := NewRussianWordRootRepeatRule(nil)
	// Java example: абрикос … абрикосный
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Абрикос рос в саду."),
		languagetool.AnalyzePlain("У меня на столе стоит абрикосный сок."),
	})
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "однокоренные")
}
