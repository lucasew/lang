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
