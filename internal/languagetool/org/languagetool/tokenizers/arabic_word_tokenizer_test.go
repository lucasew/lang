package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicWordTokenizer(t *testing.T) {
	w := NewArabicWordTokenizer()
	require.Contains(t, w.GetTokenizingCharacters(), "،")
	require.Contains(t, w.GetTokenizingCharacters(), "؟")
	require.Contains(t, w.GetTokenizingCharacters(), "؛")
	require.Contains(t, w.GetTokenizingCharacters(), "-")
	toks := w.Tokenize("مرحبا، عالم؟")
	require.Contains(t, toks, "،")
	require.Contains(t, toks, "؟")
	require.Contains(t, toks, "مرحبا")
	require.Contains(t, toks, "عالم")
	// Java ArabicWordTokenizer: glued Arabic comma becomes its own token
	// so CommaWhitespaceRule can see ، as prevToken without invent surface scan.
	glued := w.Tokenize("هذه،جملة")
	require.Equal(t, []string{"هذه", "،", "جملة"}, glued)
	// leading comma
	require.Equal(t, []string{"،", "هذه"}, w.Tokenize("،هذه"))
}
