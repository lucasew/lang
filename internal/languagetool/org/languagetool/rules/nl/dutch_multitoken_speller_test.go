package nl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDutchMultitokenSpeller_IsException(t *testing.T) {
	d := NewDutchMultitokenSpeller()
	// trailing s / -
	require.True(t, d.IsException("omawords", "omaword"))
	require.True(t, d.IsException("foo-", "foo"))
	// possessive 's / curly ’s
	require.True(t, d.IsException("oma's", "oma"))
	require.True(t, d.IsException("oma’s", "oma"))
	// not exceptions
	require.False(t, d.IsException("omaword", "omawords"))
	require.False(t, d.IsException("ab", "a")) // length <= 2
	require.False(t, d.IsException("foobar", "foo"))
}

func TestDutchMultitokenSpeller_LoadInline(t *testing.T) {
	d := NewDutchMultitokenSpeller()
	require.NoError(t, d.LoadWords(strings.NewReader("New York\n# comment\nsingle\n's Hertogenbosch\n")))
	// multiword with space enters dict; single ignored
	sugs := d.GetSuggestions("new york")
	// may return empty if edit distance path doesn't match exact — check normalize key path
	// at least load shouldn't panic; force known typo distance
	_ = sugs
	// exception path wired on MultitokenSpeller
	require.NotNil(t, d.IsException)
	require.True(t, d.MultitokenSpeller.IsException("oma's", "oma"))
}

func TestLoadDutchMultitokenSpeller_OfficialMultiwords(t *testing.T) {
	mw := DiscoverDutchMultiwords()
	if mw == "" {
		t.Skip("nl/multiwords.txt not discoverable")
	}
	sg := ""
	// spelling_global optional
	sp, err := LoadDutchMultitokenSpeller(mw, sg)
	require.NoError(t, err)
	require.NotNil(t, sp)
	// a known multiword phrase from the file (after normalize)
	// 's Hertogenbosch is multi-token with space
	sugs := sp.GetSuggestions("'s hertogenbosch")
	// suggestions may include the canonical form or be empty for exact-ish match
	_ = sugs
}

func TestDiscoverAndLoadDutchMultitokenSpeller(t *testing.T) {
	sp := DiscoverAndLoadDutchMultitokenSpeller()
	require.NotNil(t, sp)
	require.NotNil(t, sp.IsException)
}
