package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSentenceWithManRule(t *testing.T) {
	rule := NewSentenceWithManRule(nil)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Man sollte das vermeiden."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Er sollte das vermeiden."))))
}

func TestSentenceWithModalVerbRule(t *testing.T) {
	rule := NewSentenceWithModalVerbRule(nil)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Er muss das erledigen."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Er erledigt das."))))
}

func TestPassiveSentenceRule(t *testing.T) {
	rule := NewPassiveSentenceRule(nil)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Das Haus wird gebaut."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Er baut das Haus."))))
}
