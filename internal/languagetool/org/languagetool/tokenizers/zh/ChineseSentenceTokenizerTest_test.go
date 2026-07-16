package zh

// Twin of ChineseSentenceTokenizerTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChineseSentenceTokenizer_Tokenize(t *testing.T) {
	tok := NewChineseSentenceTokenizer()
	got := tok.Tokenize("你好。世界！")
	require.GreaterOrEqual(t, len(got), 2)
	require.Contains(t, got[0], "你好")
}

func TestChineseSentenceTokenizer_Tokenize2(t *testing.T) {
	tok := NewChineseSentenceTokenizer()
	got := tok.Tokenize("Hello. World.")
	require.NotEmpty(t, got)
}

func TestChineseSentenceTokenizer_TokenizeWithSpaces(t *testing.T) {
	tok := NewChineseSentenceTokenizer()
	got := tok.Tokenize("第一句。 第二句。")
	require.GreaterOrEqual(t, len(got), 2)
}
