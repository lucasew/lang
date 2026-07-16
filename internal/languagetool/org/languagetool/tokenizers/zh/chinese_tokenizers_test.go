package zh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChineseWordTokenizer(t *testing.T) {
	toks := NewChineseWordTokenizer().Tokenize("你好world")
	require.Contains(t, toks, "你")
	require.Contains(t, toks, "好")
	require.Contains(t, toks, "world")
}

func TestChineseSentenceTokenizer(t *testing.T) {
	sents := NewChineseSentenceTokenizer().Tokenize("你好。世界！")
	require.Len(t, sents, 2)
}
