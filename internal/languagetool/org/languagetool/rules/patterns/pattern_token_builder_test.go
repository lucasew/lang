package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatternTokenBuilder_Helpers(t *testing.T) {
	pt := Token("hello")
	require.Equal(t, "hello", pt.Token)
	require.False(t, pt.CaseSensitive)
	require.False(t, pt.Regexp)

	cs := CsToken("Hello")
	require.True(t, cs.CaseSensitive)

	re := TokenRegex("a|b")
	require.True(t, re.Regexp)

	pos := Pos("NN")
	require.NotNil(t, pos.Pos)
	require.Equal(t, "NN", pos.Pos.PosTag)

	pr := PosRegex("N.*")
	require.True(t, pr.Pos.Regexp)

	neg := NewPatternTokenBuilder().Token("x").Negate().Min(0).Max(2).SetSkip(1).Mark(false).Build()
	require.True(t, neg.Negation)
	require.Equal(t, 0, neg.MinOccurrence)
	require.Equal(t, 2, neg.MaxOccurrence)
	require.Equal(t, 1, neg.SkipNext)
	require.False(t, neg.InsideMarker)

	require.Panics(t, func() {
		NewPatternTokenBuilder().Token("x").Min(3).Max(1).Build()
	})
}

func TestEquivalenceTypeLocator(t *testing.T) {
	a := NewEquivalenceTypeLocator("num", "sg")
	b := NewEquivalenceTypeLocator("num", "sg")
	c := NewEquivalenceTypeLocator("num", "pl")
	require.True(t, a.Equal(b))
	require.False(t, a.Equal(c))
}
