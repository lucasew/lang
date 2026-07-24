package pl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// PolishSRXSentenceTokenizer ports tokenizers.pl.PolishSRXSentenceTokenizer.
type PolishSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewPolishSRXSentenceTokenizer() *PolishSRXSentenceTokenizer {
	return &PolishSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("pl"),
	}
}
