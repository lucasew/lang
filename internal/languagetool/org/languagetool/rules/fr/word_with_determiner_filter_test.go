package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWordWithDeterminerFilter(t *testing.T) {
	f := NewWordWithDeterminerFilter()
	require.True(t, f.IsExceptionDeterminer("nouvels"))
	require.False(t, f.IsExceptionDeterminer("les"))
	require.True(t, f.MatchesDetPOS("D def m s"))
	require.True(t, f.MatchesWordPOS("N m s"))
	require.Equal(t, "[NZ] ", f.NounAdjPrefix(true, false))
	require.Equal(t, "J ", f.NounAdjPrefix(false, true))
	require.Equal(t, "[ZNJ] ", f.NounAdjPrefix(true, true))
}
