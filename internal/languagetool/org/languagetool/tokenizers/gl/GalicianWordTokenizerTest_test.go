package gl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Unit tests for GalicianWordTokenizer (no dedicated Java WordTokenizerTest).

func tokStr(t []string) string { return "[" + strings.Join(t, ", ") + "]" }

func TestGalicianWordTokenizer_Tokenize(t *testing.T) {
	w := NewGalicianWordTokenizer()
	// decimal comma
	tokens := w.Tokenize("3,14")
	require.Equal(t, 1, len(tokens), tokStr(tokens))
	// dotted number
	tokens = w.Tokenize("1.234,56")
	require.Equal(t, 1, len(tokens), tokStr(tokens))
	// basic split
	tokens = w.Tokenize("Ola, mundo!")
	require.Equal(t, "[Ola, ,,  , mundo, !]", tokStr(tokens))
	// hyphen splits (in SPLIT_CHARS)
	tokens = w.Tokenize("pre-escolar")
	require.Equal(t, "[pre, -, escolar]", tokStr(tokens))
}
