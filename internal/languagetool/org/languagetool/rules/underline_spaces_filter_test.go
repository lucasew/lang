package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnderlineSpacesFilter(t *testing.T) {
	f := NewUnderlineSpacesFilter()
	s := "a b c"
	// "b" is at index 2..3
	from, to := f.Expand(s, 2, 3, "before")
	require.Equal(t, 1, from)
	require.Equal(t, 3, to)
	from, to = f.Expand(s, 2, 3, "after")
	require.Equal(t, 2, from)
	require.Equal(t, 4, to)
	from, to = f.Expand(s, 2, 3, "both")
	require.Equal(t, 1, from)
	require.Equal(t, 4, to)
	// no whitespace at edges
	from, to = f.Expand("abc", 1, 2, "both")
	require.Equal(t, 1, from)
	require.Equal(t, 2, to)
}
