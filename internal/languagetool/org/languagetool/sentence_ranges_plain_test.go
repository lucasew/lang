package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlainSentenceRanges(t *testing.T) {
	text := "Hello world. Next sentence!"
	rs := PlainSentenceRanges(text, "en")
	require.GreaterOrEqual(t, len(rs), 2)
	require.Equal(t, 0, rs[0].FromPos)
	require.Greater(t, rs[0].ToPos, 0)
	// ranges cover non-overlapping progressive spans
	require.GreaterOrEqual(t, rs[1].FromPos, rs[0].ToPos-1)
	// last ends within text
	require.LessOrEqual(t, rs[len(rs)-1].ToPos, len(text))
}
