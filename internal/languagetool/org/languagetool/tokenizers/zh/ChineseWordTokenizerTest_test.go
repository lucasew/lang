package zh

// Twin of ChineseWordTokenizerTest — HanLP deferred; soft-lexicon + char smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChineseWordTokenizer_Tokenize(t *testing.T) {
	tok := NewChineseWordTokenizer()
	tok.Segment = func(text string) []string { return []string{"你好", "世界"} }
	require.Equal(t, []string{"你好", "世界"}, tok.Tokenize("你好世界"))

	// Latin runs stay whole; unknown Han falls back to single chars.
	got := NewChineseWordTokenizer().Tokenize("甲乙world")
	require.Contains(t, got, "world")
	require.Contains(t, got, "甲")
	require.Contains(t, got, "乙")
}
