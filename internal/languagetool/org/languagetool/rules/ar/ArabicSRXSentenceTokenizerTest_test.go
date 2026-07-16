package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func TestArabicSRXSentenceTokenizer_Test(t *testing.T) {
	// Generic SRX for ar short code (language-specific SRX resources deferred).
	tok := tokenizers.NewSRXSentenceTokenizer("ar")
	sents := tok.Tokenize("مرحبا. كيف حالك؟")
	require.NotEmpty(t, sents)
	// Word tokenizer Arabic punctuation
	w := tokenizers.NewArabicWordTokenizer()
	words := w.Tokenize("مرحبا، كيف؟")
	require.NotEmpty(t, words)
}
