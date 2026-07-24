package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdjustLocalMatchPos(t *testing.T) {
	// sentence "Hello world", match "world" at 6..11, prior charCount 10, line 1, col 0
	m := LocalMatch{FromPos: 6, ToPos: 11, RuleID: "R", Message: "msg"}
	adj := AdjustLocalMatchPos(m, 10, 0, 1, "Hello world", nil)
	require.Equal(t, 16, adj.FromPos)
	require.Equal(t, 21, adj.ToPos)
	require.Equal(t, 6, adj.FromPosSentence)
	require.Equal(t, 11, adj.ToPosSentence)
	require.Equal(t, 1, adj.Line)
	require.Equal(t, 1, adj.EndLine)
	require.Equal(t, 6, adj.Column)
	require.Equal(t, 11, adj.EndColumn)
	require.Equal(t, 16, adj.PatternFromPos)
	require.Equal(t, 21, adj.PatternToPos)
}

func TestAdjustLocalMatchPos_Newline(t *testing.T) {
	m := LocalMatch{FromPos: 3, ToPos: 6, RuleID: "R"}
	adj := AdjustLocalMatchPos(m, 0, 5, 2, "ab\ncde", nil)
	require.Equal(t, 1, adj.Column)
	require.Equal(t, 4, adj.EndColumn)
	require.Equal(t, 3, adj.Line)
	require.Equal(t, 3, adj.EndLine)
}
