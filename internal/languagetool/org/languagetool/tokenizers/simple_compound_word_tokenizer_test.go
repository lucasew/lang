package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleCompoundWordTokenizer(t *testing.T) {
	tok := NewSimpleCompoundWordTokenizer()
	require.Equal(t, []string{"well", "known"}, tok.Tokenize("well-known"))
	require.Equal(t, []string{"hello"}, tok.Tokenize("hello"))
}
