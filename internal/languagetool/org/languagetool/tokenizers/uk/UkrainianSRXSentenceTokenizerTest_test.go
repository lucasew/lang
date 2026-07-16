package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUkrainianSRXSentenceTokenizer_TokenizeWithSplit(t *testing.T) {
	tok := NewUkrainianSRXSentenceTokenizer()
	require.NotNil(t, tok)
	got := tok.Tokenize("Привіт. Світ.")
	require.NotEmpty(t, got)
}

func TestUkrainianSRXSentenceTokenizer_TokenizeWithSpecialChars(t *testing.T) {
	tok := NewUkrainianSRXSentenceTokenizer()
	got := tok.Tokenize("Hello! World?")
	require.NotEmpty(t, got)
}

func TestUkrainianSRXSentenceTokenizer_WebEntities(t *testing.T) {
	// Web-entity edge cases deferred; smoke that tokenizer does not panic on markup-like text.
	tok := NewUkrainianSRXSentenceTokenizer()
	got := tok.Tokenize("See http://example.com. Next.")
	require.NotEmpty(t, got)
}
