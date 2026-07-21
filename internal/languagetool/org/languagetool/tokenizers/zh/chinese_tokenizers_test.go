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

func TestChineseSentenceTokenizer_CharacterIsWhitespace(t *testing.T) {
	st := NewChineseSentenceTokenizer()
	// Java Character.isWhitespace: regular space is whitespace token;
	// NBSP (U+00A0) is NOT — stays inside non-whitespace chunk (Go unicode.IsSpace would wrong).
	parts := st.Tokenize("你 好")
	require.Equal(t, []string{"你", " ", "好"}, parts)

	partsNBSP := st.Tokenize("你\u00A0好")
	// NBSP is not Character.isWhitespace → single non-whitespace chunk (no sentence seps)
	require.Equal(t, []string{"你\u00A0好"}, partsNBSP)
}
