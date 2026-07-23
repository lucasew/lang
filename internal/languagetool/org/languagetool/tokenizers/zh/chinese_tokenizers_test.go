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

func TestChineseSentenceTokenizer_CharacterIsWhitespace(t *testing.T) {
	st := NewChineseSentenceTokenizer()
	// Java Character.isWhitespace: regular space is a whitespace token of its own.
	parts := st.Tokenize("你 好")
	require.Equal(t, []string{"你", " ", "好"}, parts)

	// NBSP is NOT Character.isWhitespace, so it stays in the non-whitespace chunk
	// and is passed to SentencesUtil, which treats U+00A0 as a sentence separator
	// (lookupswitch case 160). insertIntoList uses String.trim which does NOT
	// strip NBSP, so the first sentence keeps the trailing NBSP.
	partsNBSP := st.Tokenize("你\u00A0好")
	require.Equal(t, []string{"你\u00A0", "好"}, partsNBSP)
}
