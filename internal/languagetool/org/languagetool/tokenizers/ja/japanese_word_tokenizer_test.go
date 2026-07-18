package ja

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJapaneseWordTokenizer(t *testing.T) {
	toks := NewJapaneseWordTokenizer().Tokenize("日本語ABC")
	require.NotEmpty(t, toks)
	// Latin run stays one token (encoded with POS).
	var surfaces []string
	for _, e := range toks {
		parts := splitEncoded(e)
		surfaces = append(surfaces, parts[0])
	}
	require.Contains(t, surfaces, "ABC")
	// Rare kanji still yields a token
	toks2 := NewJapaneseWordTokenizer().Tokenize("𩸽")
	require.NotEmpty(t, toks2)
}

func splitEncoded(e string) []string {
	// surface may contain no spaces; POS and lemma are last two fields after first two spaces? 
	// Java format: surface + " " + POS + " " + basic — surface has no spaces.
	// Use first space and last space carefully: POS can have no spaces (hyphenated).
	i := indexByte(e, ' ')
	if i < 0 {
		return []string{e}
	}
	rest := e[i+1:]
	j := indexByte(rest, ' ')
	if j < 0 {
		return []string{e[:i], rest}
	}
	return []string{e[:i], rest[:j], rest[j+1:]}
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
