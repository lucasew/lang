package spelling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpellingCheckRule(t *testing.T) {
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN", "spelling", "en")
	r.IsMisspelled = func(word string) bool { return word == "mispeled" }
	r.AddIgnoreWords("LanguageTool")
	require.True(t, r.AcceptWord("ok"))
	require.False(t, r.AcceptWord("mispeled"))
	require.True(t, r.AcceptWord("LanguageTool"))
	require.Equal(t, LanguageTool, "LanguageTool")
	require.Equal(t, float32(0.99), HighConfidence)
}
