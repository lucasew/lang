package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLineExpander(t *testing.T) {
	e := NewLineExpander()
	require.Equal(t, []string{"foo", "foos"}, e.ExpandLine("foo/S"))
	require.Equal(t, []string{"bar", "barn"}, e.ExpandLine("bar/N"))
	require.Contains(t, e.ExpandLine("weiter_gehen"), "weitergehen")
	require.Equal(t, []string{"plain"}, e.ExpandLine("plain"))
}
