package hunspell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileHunspellDictionary(t *testing.T) {
	dir := t.TempDir()
	dic := filepath.Join(dir, "test.dic")
	aff := filepath.Join(dir, "test.aff")
	require.NoError(t, os.WriteFile(dic, []byte("3\nhello\nworld/S\ncolour\n"), 0o644))
	require.NoError(t, os.WriteFile(aff, []byte("SET UTF-8\n"), 0o644))

	d, err := NewFileHunspellDictionary(dic, aff, false)
	require.NoError(t, err)
	require.True(t, d.Spell("hello"))
	require.True(t, d.Spell("world"))
	require.True(t, d.Spell("colour"))
	require.False(t, d.Spell("helo"))
	require.NoError(t, d.Close())
}
