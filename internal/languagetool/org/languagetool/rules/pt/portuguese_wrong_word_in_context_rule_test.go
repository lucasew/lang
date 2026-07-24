package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseWrongWordInContextRule(t *testing.T) {
	rule := NewPortugueseWrongWordInContextRule(nil)
	require.Equal(t, "PORTUGUESE_WRONG_WORD_IN_CONTEXT", rule.GetID())
	// Java example: infringiu danos → infligiu
	matches := rule.Match(languagetool.AnalyzePlain("O acidente infringiu grandes danos."))
	require.NotEmpty(t, matches)
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "inflig")
}
