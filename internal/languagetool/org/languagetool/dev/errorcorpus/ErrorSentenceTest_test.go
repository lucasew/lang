package errorcorpus

// Twin of ErrorSentenceTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorSentence_HasErrorCoveredByMatch(t *testing.T) {
	s := NewErrorSentence("this is an test", []Error{{StartPos: 8, EndPos: 10}})
	require.True(t, s.HasErrorCoveredByMatch(MatchSpan{8, 10}))
	require.True(t, s.HasErrorCoveredByMatch(MatchSpan{8, 12}))
	require.True(t, s.HasErrorCoveredByMatch(MatchSpan{7, 10}))
	require.True(t, s.HasErrorCoveredByMatch(MatchSpan{7, 11}))
	require.False(t, s.HasErrorCoveredByMatch(MatchSpan{9, 10}))
	require.False(t, s.HasErrorCoveredByMatch(MatchSpan{8, 9}))
}

func TestErrorSentence_HasErrorOverlappingWithMatch(t *testing.T) {
	s := NewErrorSentence("this is an test", []Error{{StartPos: 8, EndPos: 10}})
	require.True(t, s.HasErrorOverlappingWithMatch(MatchSpan{8, 10}))
	require.True(t, s.HasErrorOverlappingWithMatch(MatchSpan{8, 12}))
	require.True(t, s.HasErrorOverlappingWithMatch(MatchSpan{7, 10}))
	require.True(t, s.HasErrorOverlappingWithMatch(MatchSpan{7, 11}))
	require.True(t, s.HasErrorOverlappingWithMatch(MatchSpan{9, 10}))
	require.True(t, s.HasErrorOverlappingWithMatch(MatchSpan{8, 9}))
	require.True(t, s.HasErrorOverlappingWithMatch(MatchSpan{6, 8}))
	require.True(t, s.HasErrorOverlappingWithMatch(MatchSpan{10, 12}))
	require.False(t, s.HasErrorOverlappingWithMatch(MatchSpan{6, 7}))
	require.False(t, s.HasErrorOverlappingWithMatch(MatchSpan{11, 13}))
}
