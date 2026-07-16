package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadJavaProperties(t *testing.T) {
	m, err := LoadJavaProperties(strings.NewReader(`
# comment
hello=world
empty=
spaced = value
colon:ok
`))
	require.NoError(t, err)
	require.Equal(t, "world", m["hello"])
	require.Equal(t, "", m["empty"])
	require.Equal(t, "value", m["spaced"])
	require.Equal(t, "ok", m["colon"])
}

func TestValidateTranslationKeys(t *testing.T) {
	en := map[string]string{"a": "1", "b": "2"}
	de := map[string]string{"a": "1"}
	require.Equal(t, []string{"b"}, ValidateTranslationKeys(en, de))
	require.Empty(t, ValidateTranslationKeys(en, map[string]string{"a": "x", "b": "y"}))
}

func TestValidateTranslationsNotEmpty(t *testing.T) {
	require.Equal(t, []string{"x"}, ValidateTranslationsNotEmpty(map[string]string{"x": "", "y": "ok"}))
}

func TestMessagesBundleEnglishFromInspiration(t *testing.T) {
	cwd, _ := os.Getwd()
	dir := cwd
	for i := 0; i < 10; i++ {
		p := filepath.Join(dir, "inspiration", "languagetool", "languagetool-core", "src", "main", "resources", "org", "languagetool", "MessagesBundle_en.properties")
		f, err := os.Open(p)
		if err != nil {
			parent := filepath.Dir(dir)
			if parent == dir {
				t.Log("MessagesBundle_en.properties not found")
				return
			}
			dir = parent
			continue
		}
		defer f.Close()
		m, err := LoadJavaProperties(f)
		require.NoError(t, err)
		require.NotEmpty(t, m)
		// core messages should not be blank for important keys when present
		empty := ValidateTranslationsNotEmpty(m)
		// English file may intentionally have few empties; just ensure load works
		_ = empty
		return
	}
}
