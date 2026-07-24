package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenizer_WordTokenizerImplements(t *testing.T) {
	var tok Tokenizer = NewWordTokenizer()
	parts := tok.Tokenize("hello world")
	require.Contains(t, parts, "hello")
	require.Contains(t, parts, "world")
}

func TestFuncTokenizer(t *testing.T) {
	var c CompoundWordTokenizer = FuncTokenizer(func(text string) []string {
		return []string{"a", "b"}
	})
	require.Equal(t, []string{"a", "b"}, c.Tokenize("x"))
}
