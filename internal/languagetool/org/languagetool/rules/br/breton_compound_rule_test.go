package br

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestBretonCompoundRule(t *testing.T) {
	rule := NewBretonCompoundRule(nil)
	require.Equal(t, "BR_COMPOUNDS", rule.GetID())
	// Java example: alc'hweder gwez → alc'hweder-gwez
	matches := rule.Match(languagetool.AnalyzePlain("un alc'hweder gwez e-kerzh an dibenn-sizhun"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "alc'hweder-gwez")
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("un alc'hweder-gwez e-kerzh"))))
}
