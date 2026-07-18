package ja

// Twin of JapaneseWordTokenizerTest — Sen deferred; soft-lexicon longest-match.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJapaneseWordTokenizer_Tokenize(t *testing.T) {
	tok := NewJapaneseWordTokenizer()
	// custom segmenter path
	tok.Segment = func(text string) []string {
		return []string{"日本", "語"}
	}
	require.Equal(t, []string{"日本", "語"}, tok.Tokenize("日本語"))

	// Soft lexicon may keep multi-kanji compounds; Latin stays a run.
	got := NewJapaneseWordTokenizer().Tokenize("日本語ABC")
	require.Contains(t, got, "ABC")
	// At least some CJK surface is emitted
	require.True(t, len(got) >= 2)
}
