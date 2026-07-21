package morfologik

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// prepareA keeps "keep-a" lines only; prepareB keeps "keep-b".
func prepareKeepA(line string) []string {
	if line == "keep-a" {
		return []string{"keep-a"}
	}
	return []string{""}
}
func prepareKeepB(line string) []string {
	if line == "keep-b" {
		return []string{"keep-b"}
	}
	return []string{""}
}

// Twin risk: shared path + different prepareLine must not share cache (Java per-Language).
func TestPlainTextAcceptCache_SeparatePrepareLine(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "spelling_global.txt")
	require.NoError(t, os.WriteFile(p, []byte("keep-a\nkeep-b\n"), 0o644))

	a := loadPlainTextAcceptCached(p, prepareKeepA)
	b := loadPlainTextAcceptCached(p, prepareKeepB)
	require.Equal(t, []string{"keep-a"}, a)
	require.Equal(t, []string{"keep-b"}, b)
	// second load from cache still distinct
	require.Equal(t, []string{"keep-a"}, loadPlainTextAcceptCached(p, prepareKeepA))
	require.Equal(t, []string{"keep-b"}, loadPlainTextAcceptCached(p, prepareKeepB))
}
