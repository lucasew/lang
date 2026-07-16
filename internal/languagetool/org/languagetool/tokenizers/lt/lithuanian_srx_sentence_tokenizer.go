package lt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// LithuanianSRXSentenceTokenizer ports tokenizers.lt.LithuanianSRXSentenceTokenizer.
type LithuanianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewLithuanianSRXSentenceTokenizer() *LithuanianSRXSentenceTokenizer {
	return &LithuanianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("lt"),
	}
}
