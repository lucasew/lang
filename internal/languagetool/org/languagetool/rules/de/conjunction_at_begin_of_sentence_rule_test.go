package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestConjunctionAtBeginOfSentenceRule(t *testing.T) {
	rule := NewConjunctionAtBeginOfSentenceRule(nil)
	// "Und" at start
	matches := rule.Match(languagetool.AnalyzePlain("Und dann ging er weg."))
	require.Equal(t, 1, len(matches))
	// "Wie" exception
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Wie geht es dir?"))))
	// not a conjunction
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Er ging dann weg."))))
}
