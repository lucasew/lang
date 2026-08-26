package commandline

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandInputPaths_Recursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "b.md"), []byte("y"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "c.bin"), []byte{0, 1}, 0o644))
	// .bin not texty — skipped
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".hidden"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".hidden", "secret.txt"), []byte("z"), 0o644))

	files, err := ExpandInputPaths([]string{dir}, true)
	require.NoError(t, err)
	require.Contains(t, files, filepath.Join(dir, "a.txt"))
	require.Contains(t, files, filepath.Join(sub, "b.md"))
	require.NotContains(t, files, filepath.Join(sub, "c.bin"))
	// hidden dir skipped
	for _, f := range files {
		require.NotContains(t, f, ".hidden")
	}
}

func TestExpandInputPaths_NoRecursiveDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644))
	files, err := ExpandInputPaths([]string{dir}, false)
	require.NoError(t, err)
	// directory without -r yields empty → treated as stdin placeholder only if nothing
	// our impl: skip dir contents → empty → return [""]
	require.Equal(t, []string{""}, files)
}

func TestRunWithIO_RecursiveLint(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "bad.txt"), []byte("This is an test."), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "ok.txt"), []byte("All good here."), 0o644))

	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--lint", "-r", dir}, DefaultCoreHooks(), &out, &errb)
	require.Equal(t, 1, code, errb.String())
	require.Contains(t, out.String(), "bad.txt")
	require.Contains(t, out.String(), "EN_A_VS_AN")
}
