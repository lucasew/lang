package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSymSpellRule(t *testing.T) {
	r := NewSymSpellRule("SYMSPELL_RULE", "en")
	r.AddWords("hello", "world")
	require.True(t, r.isMisspelled("helo"))
	require.False(t, r.isMisspelled("hello"))
	sent := languagetool.AnalyzePlain("hello helo")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}
