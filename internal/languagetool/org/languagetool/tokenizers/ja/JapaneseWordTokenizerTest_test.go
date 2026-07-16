package ja

// Twin of JapaneseWordTokenizerTest — Kuromoji deferred; script-run fallback smokes.
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

	// fallback: kanji split + latin run
	got := NewJapaneseWordTokenizer().Tokenize("日本語ABC")
	require.Contains(t, got, "日")
	require.Contains(t, got, "本")
	require.Contains(t, got, "語")
	require.Contains(t, got, "ABC")
}
