package en

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareLineForSpeller(t *testing.T) {
	require.Equal(t, []string{"New York"}, PrepareLineForSpeller("New York\tNNP"))
	require.Equal(t, []string{"big blue"}, PrepareLineForSpeller("big blue\tJJ"))
	require.Equal(t, []string{""}, PrepareLineForSpeller("foo bar\tVB"))
	require.Equal(t, []string{"P. Sherman 42 Wallaby Way"}, PrepareLineForSpeller("P. Sherman 42 Wallaby Way"))
	require.Equal(t, []string{""}, PrepareLineForSpeller("foo+bar\tNN"))
}

func TestLoadEnglishMultitokenSpeller_OfficialMultiwords(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// rules/en → languagetool → org → languagetool → internal → repo root (5 levels up from en)
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..", ".."))
	mw := filepath.Join(root, "inspiration", "languagetool", "languagetool-language-modules", "en",
		"src", "main", "resources", "org", "languagetool", "resource", "en", "multiwords.txt")
	if st, err := os.Stat(mw); err != nil || st.IsDir() {
		t.Skipf("official multiwords missing: %s", mw)
	}
	sp, err := LoadEnglishMultitokenSpeller(mw, "")
	require.NoError(t, err)
	require.NotNil(t, sp)
	// exact known multiword → stopSearching (no suggestions)
	require.Empty(t, sp.GetSuggestions("Taj Mahal"))
	require.Empty(t, sp.GetSuggestions("status quo"))
	// typo against known entry
	sugg := sp.GetSuggestions("Moulin Ruge")
	require.Contains(t, sugg, "Moulin Rouge")
}

func TestLoadEnglishMultitokenSpeller_Inline(t *testing.T) {
	sp := NewEnglishMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\tNNP\nfoo bar\tVB\nhello world\n")))
	require.Empty(t, sp.GetSuggestions("foo bar"), "VB line dropped by prepareLineForSpeller")
	require.Contains(t, sp.GetSuggestions("New Yrok"), "New York")
	require.Contains(t, sp.GetSuggestions("helo world"), "hello world")
}
