package ja

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Extra coverage beyond JapaneseWordTokenizerTest (not a Java twin).
func TestJapaneseWordTokenizer_mixedAndUnknown(t *testing.T) {
	toks := NewJapaneseWordTokenizer().Tokenize("日本語ABC")
	require.NotEmpty(t, toks)
	var surfaces []string
	for _, e := range toks {
		// encoded: surface + " " + POS + " " + basic — surface has no spaces
		i := strings.IndexByte(e, ' ')
		require.Greater(t, i, 0)
		surfaces = append(surfaces, e[:i])
	}
	require.Contains(t, surfaces, "ABC")

	toks2 := NewJapaneseWordTokenizer().Tokenize("𩸽")
	require.NotEmpty(t, toks2)
	require.True(t, strings.HasPrefix(toks2[0], "𩸽 "))
}

func TestJapaneseWordTokenizer_empty(t *testing.T) {
	// Java analyze of empty yields empty list (not null).
	got := NewJapaneseWordTokenizer().Tokenize("")
	require.Empty(t, got)
	require.NotNil(t, got) // empty slice, same as Java empty ArrayList
}
