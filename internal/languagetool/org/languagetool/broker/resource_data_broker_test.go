package broker

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestMapResourceDataBroker(t *testing.T) {
	m := NewMapResourceDataBroker()
	m.Resource["en/words.txt"] = "a\nb\n"
	require.True(t, m.ResourceExists("en/words.txt"))
	require.False(t, m.RuleFileExists("en/words.txt"))
	lines, err := m.GetFromResourceDirAsLines("en/words.txt")
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, lines)
}

func TestFSResourceDataBroker(t *testing.T) {
	fsys := fstest.MapFS{
		"res/xx/file.txt": &fstest.MapFile{Data: []byte("hello\nworld\n")},
	}
	b := NewFSResourceDataBroker(fsys, "res", "rules")
	require.True(t, b.ResourceExists("xx/file.txt"))
	lines, err := b.GetFromResourceDirAsLines("xx/file.txt")
	require.NoError(t, err)
	require.Equal(t, []string{"hello", "world"}, lines)
}
