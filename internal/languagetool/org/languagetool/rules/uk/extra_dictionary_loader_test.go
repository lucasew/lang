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

// Twin: loadMap uses String.trim (not Unicode) + split(" ").
func TestLoadMap_JavaTrimAndSpaceSplit(t *testing.T) {
	// NBSP after key is not trimmed; split(" ") keeps whole "a\u00a0b" as one field when no ASCII space.
	m, err := LoadMap(strings.NewReader("k\u00a0v\nx y z\n"))
	require.NoError(t, err)
	require.Equal(t, "", m["k\u00a0v"]) // single field after trim of ASCII edges
	require.Equal(t, "y", m["x"])       // only parts[1], not rest
}

func TestLoadSpacedLists(t *testing.T) {
	m, err := LoadSpacedLists(strings.NewReader("key a b|c\n#skip\n"))
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, m["key"])
}

// Twin: split(" |\\|") keeps empty mid-fields (Java String.split limit 0).
func TestSplitSpaceOrPipe_EmptyMid(t *testing.T) {
	require.Equal(t, []string{"a", "", "b"}, splitSpaceOrPipe("a  b"))
	require.Equal(t, []string{"a", "b", "c"}, splitSpaceOrPipe("a b|c"))
}

func TestLoadLists(t *testing.T) {
	m, err := LoadLists(strings.NewReader("word = x|y\nother=z\n"))
	require.NoError(t, err)
	require.Equal(t, []string{"x", "y"}, m["word"])
	require.Equal(t, []string{"z"}, m["other"])
}
