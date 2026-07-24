package be

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// BelarusianSRXSentenceTokenizer ports tokenizers.be.BelarusianSRXSentenceTokenizer.
type BelarusianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewBelarusianSRXSentenceTokenizer() *BelarusianSRXSentenceTokenizer {
	return &BelarusianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("be"),
	}
}
