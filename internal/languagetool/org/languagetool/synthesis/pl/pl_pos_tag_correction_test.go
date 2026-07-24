package pl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolishSynthesizer_GetPosTagCorrection(t *testing.T) {
	s := NewPolishSynthesizer(nil)
	// no dot → identity
	require.Equal(t, "subst:sg:nom:m1", s.GetPosTagCorrection("subst:sg:nom:m1"))
	// segment with a.z style dots expands (Java PolishSynthesizer)
	got := s.GetPosTagCorrection("adj:a.b:sg")
	require.Contains(t, got, ".*a.*|.*b.*")
	require.Contains(t, got, "adj:")
	// whole-tag with letter.letter expands (Java DOT → .*|.*)
	require.Equal(t, "(.*foo.*|.*bar.*)", s.GetPosTagCorrection("foo.bar"))
	// upper-case around dot does not match .*[a-z]\.[a-z].*
	require.Equal(t, "FOO.BAR", s.GetPosTagCorrection("FOO.BAR"))
}
