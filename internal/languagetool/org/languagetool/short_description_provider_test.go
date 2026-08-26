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

func TestShortDescriptionProviderBadFormat(t *testing.T) {
	p := NewShortDescriptionProvider()
	p.LoadLines = func(path string) ([]string, error) {
		return []string{"not-tab-separated"}, nil
	}
	require.Panics(t, func() { _ = p.GetShortDescription("x", "en") })
}

func TestShortDescriptionProviderHashCommentRaw(t *testing.T) {
	// Java: startsWith("#") on raw line — " # not comment" is not a comment if trimmed would be
	p := NewShortDescriptionProvider()
	p.LoadLines = func(path string) ([]string, error) {
		return []string{" #not\tcomment style"}, nil
	}
	// has tab → 2 parts, first is " #not"
	require.Equal(t, "comment style", p.GetShortDescription(" #not", "en"))
}

func TestParseWordDefinitions(t *testing.T) {
	m, err := ParseWordDefinitions(strings.NewReader("a\tb\n"))
	require.NoError(t, err)
	require.Equal(t, "b", m["a"])
}
