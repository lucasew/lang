package eo

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// EsperantoSRXSentenceTokenizer ports tokenizers.eo.EsperantoSRXSentenceTokenizer.
type EsperantoSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewEsperantoSRXSentenceTokenizer() *EsperantoSRXSentenceTokenizer {
	return &EsperantoSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("eo"),
	}
}
