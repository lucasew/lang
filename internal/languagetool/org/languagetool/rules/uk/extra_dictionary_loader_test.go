package uk

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSet(t *testing.T) {
	m, err := LoadSet(strings.NewReader("# c\nfoo\nbar\n"))
	require.NoError(t, err)
	_, ok := m["foo"]
	require.True(t, ok)
	_, ok = m["# c"]
	require.False(t, ok)
}

func TestLoadMap(t *testing.T) {
	m, err := LoadMap(strings.NewReader("a b\nc\n"))
	require.NoError(t, err)
	require.Equal(t, "b", m["a"])
	require.Equal(t, "", m["c"])
}

func TestLoadSpacedLists(t *testing.T) {
	m, err := LoadSpacedLists(strings.NewReader("key a b|c\n#skip\n"))
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, m["key"])
}

func TestLoadLists(t *testing.T) {
	m, err := LoadLists(strings.NewReader("word = x|y\nother=z\n"))
	require.NoError(t, err)
	require.Equal(t, []string{"x", "y"}, m["word"])
	require.Equal(t, []string{"z"}, m["other"])
}
