package zh

// Twin of ChineseWordTokenizerTest — HanLP deferred; incomplete per-rune CJK + Latin runs.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChineseWordTokenizer_Tokenize(t *testing.T) {
	tok := NewChineseWordTokenizer()
	tok.Segment = func(text string) []string {
		return []string{"甲", "乙"}
	}
	got := tok.Tokenize("甲乙")
	require.Equal(t, []string{"甲/x", "乙/x"}, got)

	raw := NewChineseWordTokenizer().Tokenize("甲乙world")
	var surfaces []string
	for _, e := range raw {
		surfaces = append(surfaces, strings.SplitN(e, "/", 2)[0])
	}
	require.Contains(t, surfaces, "world")
}
