package it

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// ItalianSRXSentenceTokenizer ports tokenizers.it.ItalianSRXSentenceTokenizer.
type ItalianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewItalianSRXSentenceTokenizer() *ItalianSRXSentenceTokenizer {
	return &ItalianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("it"),
	}
}
