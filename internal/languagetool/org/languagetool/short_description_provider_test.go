package languagetool

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShortDescriptionProvider(t *testing.T) {
	p := NewShortDescriptionProvider()
	p.LoadLines = func(path string) ([]string, error) {
		require.Equal(t, "/en/word_definitions.txt", path)
		return []string{
			"# comment",
			"dog\ta domestic animal",
			"cat\ta small feline",
		}, nil
	}
	require.Equal(t, "a domestic animal", p.GetShortDescription("dog", "en"))
	require.Equal(t, "", p.GetShortDescription("wolf", "en"))
	// cached
	require.Equal(t, "a small feline", p.GetShortDescription("cat", "en"))
}

func TestParseWordDefinitions(t *testing.T) {
	m, err := ParseWordDefinitions(strings.NewReader("a\tb\n"))
	require.NoError(t, err)
	require.Equal(t, "b", m["a"])
}
