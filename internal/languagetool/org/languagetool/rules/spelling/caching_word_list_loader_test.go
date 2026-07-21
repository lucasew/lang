package spelling

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCachingWordListLoader(t *testing.T) {
	l := NewCachingWordListLoader()
	words, err := l.LoadWordsFromReader("x.txt", strings.NewReader("#c\nfoo #trail\nbar\n\n"))
	require.NoError(t, err)
	require.Equal(t, []string{"foo", "bar"}, words)
	// cached — second call with different reader still returns first
	words2, err := l.LoadWordsFromReader("x.txt", strings.NewReader("ignored\n"))
	require.NoError(t, err)
	require.Equal(t, words, words2)
	require.Equal(t, words, l.LoadWords("x.txt"))
}

// Twin: CachingWordListLoader uses String.trim (≤U+0020), not Unicode TrimSpace.
// NBSP-only suffix must remain part of the entry until comment strip; NBSP is not trimmed.
func TestParseWordListLines_JavaTrimNotUnicode(t *testing.T) {
	// "foo" + NBSP after trim of ASCII spaces stays "foo\u00a0" then no # → kept with NBSP.
	in := "  foo\u00a0\nbar #x\n"
	got, err := ParseWordListLines(strings.NewReader(in))
	require.NoError(t, err)
	require.Equal(t, []string{"foo\u00a0", "bar"}, got)
}
