package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/RemoteRuleMatchTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of RemoteRuleMatchTest.isTouchedByOneOf
func TestRemoteRuleMatch_IsTouchedByOneOf(t *testing.T) {
	// Java RuleMatch spans: [0,3) and [10,13)
	orig := []Span{{From: 0, To: 3}, {From: 10, To: 13}}

	require.True(t, matchSpan(0, 5).IsTouchedByOneOf(orig))
	require.True(t, matchSpan(2, 3).IsTouchedByOneOf(orig))

	require.False(t, matchSpan(4, 5).IsTouchedByOneOf(orig))
	require.False(t, matchSpan(4, 9).IsTouchedByOneOf(orig))

	require.True(t, matchSpan(8, 10).IsTouchedByOneOf(orig))
	require.True(t, matchSpan(10, 13).IsTouchedByOneOf(orig))
	require.True(t, matchSpan(12, 13).IsTouchedByOneOf(orig))
	require.True(t, matchSpan(12, 15).IsTouchedByOneOf(orig))
	require.True(t, matchSpan(8, 15).IsTouchedByOneOf(orig))

	require.False(t, matchSpan(14, 20).IsTouchedByOneOf(orig))
}

func matchSpan(from, to int) *RemoteRuleMatch {
	return NewRemoteRuleMatch("R1", "msg", "...", 0, from, to-from)
}
