package zh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChineseWordTokenizer(t *testing.T) {
	toks := NewChineseWordTokenizer().Tokenize("你好world")
	require.NotEmpty(t, toks)
	// Encoded HanLP-style surface/pos
	var surfaces []string
	for _, e := range toks {
		parts := strings.SplitN(e, "/", 2)
		surfaces = append(surfaces, parts[0])
		require.Contains(t, e, "/")
	}
	require.Contains(t, surfaces, "world")
}

func TestChineseSentenceTokenizer(t *testing.T) {
	st := NewChineseSentenceTokenizer()
	require.NotEmpty(t, st.Tokenize("你好。世界！"))
}
