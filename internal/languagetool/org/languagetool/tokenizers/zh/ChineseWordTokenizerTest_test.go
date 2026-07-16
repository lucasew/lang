package zh

// Twin of ChineseWordTokenizerTest — full ICTCLAS deferred; char-level smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChineseWordTokenizer_Tokenize(t *testing.T) {
	tok := NewChineseWordTokenizer()
	tok.Segment = func(text string) []string { return []string{"你好", "世界"} }
	require.Equal(t, []string{"你好", "世界"}, tok.Tokenize("你好世界"))

	got := NewChineseWordTokenizer().Tokenize("你好world")
	require.Equal(t, []string{"你", "好", "world"}, got)
}
