package server

// Twin of ConfidenceMapLoaderTest
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

func TestConfidenceMapLoader_Loading(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf-en.csv")
	require.NoError(t, os.WriteFile(path, []byte("RULE_A,0.75\n# c\nRULE_B,0.5\n"), 0o644))
	loader := NewConfidenceMapLoader()
	m, err := loader.Load(filepath.Join(dir, "conf-{lang}.csv"), []string{"en"})
	require.NoError(t, err)
	require.NotEmpty(t, m)
}

func TestConfidenceMapLoader_LoadingFail(t *testing.T) {
	loader := NewConfidenceMapLoader()
	// missing {lang}
	_, err := loader.Load("/tmp/conf.csv", []string{"en"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "{lang}")
	// no files → empty map error
	_, err = loader.Load(filepath.Join(t.TempDir(), "conf-{lang}.csv"), []string{"zz"})
	require.Error(t, err)
}

// Twin: Java does not trim line/fields — key is parts[0] as-is after split(",").
func TestConfidenceMapLoader_NoTrim(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf-en.csv")
	// trailing spaces on rule id kept
	require.NoError(t, os.WriteFile(path, []byte("RULE_SP ,0.1\n"), 0o644))
	loader := NewConfidenceMapLoader()
	m, err := loader.Load(filepath.Join(dir, "conf-{lang}.csv"), []string{"en"})
	require.NoError(t, err)
	// key uses parts[0] including trailing space
	_, okSpaced := m[tools.NewConfidenceKey("en", "RULE_SP ")]
	require.True(t, okSpaced)
}
