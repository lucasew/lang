package ja

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJapaneseWordTokenizer(t *testing.T) {
	toks := NewJapaneseWordTokenizer().Tokenize("日本語ABC")
	require.NotEmpty(t, toks)
	require.Contains(t, toks, "日")
	require.Contains(t, toks, "ABC")
}
