package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicWordTokenizer(t *testing.T) {
	w := NewArabicWordTokenizer()
	require.Contains(t, w.GetTokenizingCharacters(), "،")
	require.Contains(t, w.GetTokenizingCharacters(), "؟")
	toks := w.Tokenize("مرحبا، عالم؟")
	require.Contains(t, toks, "،")
	require.Contains(t, toks, "؟")
	require.Contains(t, toks, "مرحبا")
	require.Contains(t, toks, "عالم")
}
