package languagetool

// Twin of ShortDescriptionProviderTest
import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// twinWordDefinitions loads Java resource org/languagetool/resource/{lang}/word_definitions.txt
func twinWordDefinitions(t *testing.T, lang string) func(path string) ([]string, error) {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// .../internal/languagetool/org/languagetool/ → repo root is 4 levels up
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", ".."))
	// language module layout
	res := filepath.Join(root, "inspiration", "languagetool", "languagetool-language-modules", lang,
		"src", "main", "resources", "org", "languagetool", "resource", lang, "word_definitions.txt")
	if _, err := os.Stat(res); err != nil {
		t.Skipf("word_definitions.txt not available for %s: %v", lang, err)
	}
	return func(path string) ([]string, error) {
		// Java path "/{lang}/word_definitions.txt"
		want := "/" + lang + "/word_definitions.txt"
		if path != want {
			return nil, os.ErrNotExist
		}
		f, err := os.Open(res)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var lines []string
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			lines = append(lines, sc.Text())
		}
		return lines, sc.Err()
	}
}

// Twin of ShortDescriptionProviderTest.testGetShortDescription
func TestShortDescriptionProvider_GetShortDescription(t *testing.T) {
	// de-DE → short code "de"
	pDE := NewShortDescriptionProvider()
	pDE.LoadLines = twinWordDefinitions(t, "de")
	require.NotEmpty(t, pDE.GetShortDescription("fielen", "de"))
	require.Empty(t, pDE.GetShortDescription("fake-word-doesnt-exist", "de"))

	// en-US → short code "en"
	pEN := NewShortDescriptionProvider()
	pEN.LoadLines = twinWordDefinitions(t, "en")
	require.NotEmpty(t, pEN.GetShortDescription("adopting", "en"))
	require.Empty(t, pEN.GetShortDescription("fake-word-doesnt-exist", "en"))
}

func TestShortDescriptionProvider_DescriptionLength(t *testing.T) {
	// Java soft-warns when len > 45; fails only if zero descriptions across langs.
	// Twin: load EN + DE official resources and assert at least some entries, most ≤45.
	limit := 45
	count := 0
	over := 0
	for _, lang := range []string{"en", "de"} {
		p := NewShortDescriptionProvider()
		p.LoadLines = twinWordDefinitions(t, lang)
		// force init via known word
		_ = p.GetShortDescription("__force_load__", lang)
		// re-load map via private path: scan file ourselves for length check
		lines, err := twinWordDefinitions(t, lang)("/" + lang + "/word_definitions.txt")
		require.NoError(t, err)
		for _, line := range lines {
			if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
				continue
			}
			parts := strings.Split(line, "\t")
			if len(parts) != 2 {
				continue
			}
			count++
			if len(parts[1]) > limit {
				over++
			}
		}
	}
	require.Greater(t, count, 0, "No word descriptions found")
	// Java only logs WARNING for over-limit; do not fail — just ensure file is usable
	_ = over
}
