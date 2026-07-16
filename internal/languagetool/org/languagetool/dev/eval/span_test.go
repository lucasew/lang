package eval

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpan_Covers(t *testing.T) {
	a := NewSpan(0, 10)
	b := NewSpan(2, 5)
	require.True(t, a.Covers(b))
	require.False(t, b.Covers(a))
	require.True(t, a.Overlaps(b))
	require.False(t, NewSpan(0, 2).Overlaps(NewSpan(2, 4)))
}
