package ja

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJapaneseWordTokenizer(t *testing.T) {
	toks := NewJapaneseWordTokenizer().Tokenize("日本語ABC")
	require.NotEmpty(t, toks)
	// Soft lexicon may keep multi-kanji compounds (e.g. 日本); Latin stays a run.
	require.Contains(t, toks, "ABC")
	// Character fallback still applies for unknown CJK
	toks2 := NewJapaneseWordTokenizer().Tokenize("𩸽") // rare kanji unlikely in soft lex
	require.Equal(t, []string{"𩸽"}, toks2)
}
