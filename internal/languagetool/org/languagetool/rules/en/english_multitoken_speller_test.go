package en

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

func TestEnglishMultitokenResourcePaths(t *testing.T) {
	// Java: Arrays.asList("/en/multiwords.txt", "/spelling_global.txt")
	require.Equal(t, []string{"/en/multiwords.txt", "/spelling_global.txt"}, EnglishMultitokenResourcePaths)
	require.NotNil(t, INSTANCE)
	require.NotNil(t, EnglishMultitokenSpellerInstance)
	require.Same(t, INSTANCE, EnglishMultitokenSpellerInstance)
	require.NotNil(t, INSTANCE.MultitokenSpeller)
	require.NotNil(t, INSTANCE.PrepareLine)
}

func TestPrepareLineForSpeller(t *testing.T) {
	// Twin of English.prepareLineForSpeller (Java ~line 741)
	require.Equal(t, []string{"New York"}, PrepareLineForSpeller("New York\tNNP"))
	require.Equal(t, []string{"big blue"}, PrepareLineForSpeller("big blue\tJJ"))
	require.Equal(t, []string{""}, PrepareLineForSpeller("foo bar\tVB"))
	require.Equal(t, []string{"P. Sherman 42 Wallaby Way"}, PrepareLineForSpeller("P. Sherman 42 Wallaby Way"))
	require.Equal(t, []string{""}, PrepareLineForSpeller("foo+bar\tNN"))
	// NN:UN (status quo style) still starts with NN
	require.Equal(t, []string{"status quo"}, PrepareLineForSpeller("status quo\tNN:UN"))
	// comment after tag
	require.Equal(t, []string{"word"}, PrepareLineForSpeller("word\tNN#comment"))
	// language package twin must stay identical (single prepareLine implementation)
	for _, line := range []string{
		"New York\tNNP", "foo bar\tVB", "foo+bar\tNN", "plain", "word\tNN#comment",
	} {
		require.Equal(t, language.EnglishPrepareLineForSpeller(line), PrepareLineForSpeller(line), line)
	}
}

func TestLoadEnglishMultitokenSpeller_OfficialMultiwords(t *testing.T) {
	mw := DiscoverEnglishMultiwords()
	if mw == "" {
		t.Skip("official en/multiwords.txt not discoverable")
	}
	// multiwords only first (Java still loads spelling_global after; test file alone)
	sp, err := LoadEnglishMultitokenSpeller(mw, "")
	require.NoError(t, err)
	require.NotNil(t, sp)
	// exact known multiword → stopSearching (no suggestions)
	require.Empty(t, sp.GetSuggestions("Taj Mahal"))
	require.Empty(t, sp.GetSuggestions("status quo"))
	// typo against known entry (Moulin Rouge\tNNP in official multiwords.txt)
	sugg := sp.GetSuggestions("Moulin Ruge")
	require.Contains(t, sugg, "Moulin Rouge")
}

func TestLoadEnglishMultitokenSpeller_WithSpellingGlobal(t *testing.T) {
	mw := DiscoverEnglishMultiwords()
	sg := spelling.DiscoverSpellingGlobal()
	if mw == "" && sg == "" {
		t.Skip("official multiwords/spelling_global not discoverable")
	}
	sp, err := LoadEnglishMultitokenSpeller(mw, sg)
	require.NoError(t, err)
	require.NotNil(t, sp)
	// multiwords exact still stops
	if mw != "" {
		require.Empty(t, sp.GetSuggestions("Taj Mahal"))
		require.Contains(t, sp.GetSuggestions("Moulin Ruge"), "Moulin Rouge")
	}
	// spelling_global untagged multi-token lines pass prepareLine as whole line
	// (English.prepareLineForSpeller returns Collections.singletonList(line) when no tab).
	if sg != "" {
		// exact phrase from official spelling_global.txt
		require.Empty(t, sp.GetSuggestions("P. Sherman 42 Wallaby Way"))
		// typo → suggestion from spelling_global
		sugg := sp.GetSuggestions("Walt Disney Wrld")
		require.Contains(t, sugg, "Walt Disney World")
	}
}

func TestDiscoverAndLoadEnglishMultitokenSpeller_Official(t *testing.T) {
	if DiscoverEnglishMultiwords() == "" && spelling.DiscoverSpellingGlobal() == "" {
		t.Skip("official EN multitoken resources not discoverable")
	}
	sp := DiscoverAndLoadEnglishMultitokenSpeller()
	require.NotNil(t, sp)
	require.NotNil(t, sp.PrepareLine)
	// same prepareLine hook as English.prepareLineForSpeller
	require.Equal(t, []string{"New York"}, sp.PrepareLine("New York\tNNP"))
	require.Equal(t, []string{""}, sp.PrepareLine("foo bar\tVB"))

	if DiscoverEnglishMultiwords() != "" {
		require.Empty(t, sp.GetSuggestions("Taj Mahal"))
		require.Empty(t, sp.GetSuggestions("status quo"))
		require.Contains(t, sp.GetSuggestions("Moulin Ruge"), "Moulin Rouge")
	}
	if spelling.DiscoverSpellingGlobal() != "" {
		require.Empty(t, sp.GetSuggestions("P. Sherman 42 Wallaby Way"))
		require.Contains(t, sp.GetSuggestions("Walt Disney Wrld"), "Walt Disney World")
	}
}

func TestEnglishMultitokenSpeller_INSTANCE_Outcomes(t *testing.T) {
	// INSTANCE is package-init DiscoverAndLoad; skip if resources missing at init time.
	if DiscoverEnglishMultiwords() == "" {
		t.Skip("official en/multiwords.txt not discoverable (INSTANCE may be empty)")
	}
	require.NotNil(t, INSTANCE)
	require.Empty(t, INSTANCE.GetSuggestions("Taj Mahal"))
	require.Contains(t, INSTANCE.GetSuggestions("Moulin Ruge"), "Moulin Rouge")
	if spelling.DiscoverSpellingGlobal() != "" {
		require.Contains(t, INSTANCE.GetSuggestions("Walt Disney Wrld"), "Walt Disney World")
	}
}

func TestLoadEnglishMultitokenSpeller_Inline(t *testing.T) {
	sp := NewEnglishMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("New York\tNNP\nfoo bar\tVB\nhello world\n")))
	require.Empty(t, sp.GetSuggestions("foo bar"), "VB line dropped by prepareLineForSpeller")
	require.Contains(t, sp.GetSuggestions("New Yrok"), "New York")
	require.Contains(t, sp.GetSuggestions("helo world"), "hello world")
	// untagged multiword keeps full line (spelling_global style)
	require.Empty(t, sp.GetSuggestions("hello world"))
}
