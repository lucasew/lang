package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPersianWordTokenizer(t *testing.T) {
	w := NewPersianWordTokenizer()
	require.Contains(t, w.GetTokenizingCharacters(), "،")
	toks := w.Tokenize("سلام، دنیا؟")
	require.Contains(t, toks, "،")
	require.Contains(t, toks, "سلام")
}
