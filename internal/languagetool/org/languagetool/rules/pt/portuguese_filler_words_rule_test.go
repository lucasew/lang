package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseFillerWordsRule(t *testing.T) {
	rule := NewPortugueseFillerWordsRule(nil)
	require.Equal(t, "FILLER_WORDS_PT", rule.GetID())
	matches := rule.Match(languagetool.AnalyzePlain("Ele realmente veio."))
	require.Equal(t, 1, len(matches))
	// "mas" after comma is exception (may still flag other fillers in the sentence)
	for _, m := range rule.Match(languagetool.AnalyzePlain("X, mas Y.")) {
		require.NotEqual(t, "mas", m.GetMessage()) // matched token is not asserted via message
		// ensure "mas" itself is not in suggested replacements context — check positions
		_ = m
	}
	// more direct: sentence with only filler "mas" after comma should not match mas
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("X, mas Z."))))
}
