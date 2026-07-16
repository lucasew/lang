package ja

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// JapaneseSRXSentenceTokenizer ports tokenizers.ja.JapaneseSRXSentenceTokenizer.
type JapaneseSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewJapaneseSRXSentenceTokenizer() *JapaneseSRXSentenceTokenizer {
	return &JapaneseSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("ja"),
	}
}
