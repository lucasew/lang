package bigdata

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNgramCounts_IndexSentence(t *testing.T) {
	c := NewNgramCounts()
	c.IndexSentence("The cat sat.", nil)
	require.Greater(t, c.Unigram["The"], int64(0))
	require.Greater(t, c.Unigram[GoogleSentenceStart], int64(0))
	require.Greater(t, c.Unigram[GoogleSentenceEnd], int64(0))
	// bigram START The
	require.Greater(t, c.Bigram[GoogleSentenceStart+" The"], int64(0))
	// trigram present
	require.NotEmpty(t, c.Trigram)
}

func TestNgramCounts_IndexLines(t *testing.T) {
	c := NewNgramCounts()
	require.NoError(t, c.IndexLines(strings.NewReader("Hello world\nBye now\n"), nil))
	require.Greater(t, c.Unigram["Hello"], int64(0))
	require.Greater(t, c.Unigram["Bye"], int64(0))
}
