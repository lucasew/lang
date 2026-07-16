package el

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGreekWordTokenizer(t *testing.T) {
	toks := NewGreekWordTokenizer().Tokenize("Γεια σου")
	require.Contains(t, toks, "Γεια")
	require.Contains(t, toks, "σου")
}
