package be

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBelarusianWordTokenizer_Tokenize(t *testing.T) {
	w := NewBelarusianWordTokenizer()
	tokens := w.Tokenize("камп'ютар")
	require.Equal(t, 1, len(tokens))
	require.Equal(t, []string{"камп'ютар"}, tokens)

	tokens2 := w.Tokenize("Яно\rразбіваецца")
	require.Equal(t, 3, len(tokens2))
	require.Equal(t, "[Яно, \r, разбіваецца]", "["+strings.Join(tokens2, ", ")+"]")

	tokens3 := w.Tokenize("Мой адрас — address@email.com")
	require.Equal(t, 7, len(tokens3))
	require.Equal(t, "[Мой,  , адрас,  , —,  , address@email.com]", "["+strings.Join(tokens3, ", ")+"]")
}
