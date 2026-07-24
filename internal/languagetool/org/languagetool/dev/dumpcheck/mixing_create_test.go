package dumpcheck

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateMixingSentenceSource_Files(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "sample.txt")
	tat := filepath.Join(dir, "tatoeba-en.txt")
	require.NoError(t, os.WriteFile(plain, []byte("Plain sample sentence is long enough here.\n"), 0o644))
	require.NoError(t, os.WriteFile(tat, []byte("1\teng\tTatoeba sample sentence is long enough here.\n"), 0o644))

	mix, err := CreateMixingSentenceSource([]string{plain, tat}, "en")
	require.NoError(t, err)
	require.True(t, mix.HasNext())
	s1, err := mix.Next()
	require.NoError(t, err)
	s2, err := mix.Next()
	require.NoError(t, err)
	// alternate plain then tatoeba
	require.Contains(t, s1.GetText()+s2.GetText(), "Plain")
	require.Contains(t, s1.GetText()+s2.GetText(), "Tatoeba")
}

func TestCreateMixingSentenceSource_Unknown(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "file.xyz")
	require.NoError(t, os.WriteFile(bad, []byte("x"), 0o644))
	_, err := CreateMixingSentenceSource([]string{bad}, "en")
	require.Error(t, err)
}
