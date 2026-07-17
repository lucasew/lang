package bigdata

// Twin of CommonCrawlToNgramTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommonCrawlToNgram_NoTests(t *testing.T) {
	c := NewNgramCounts()
	c.IndexSentence("ngram test line", nil)
	require.NotEmpty(t, c.Unigram)
}
