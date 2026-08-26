package uk

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// UkrainianSRXSentenceTokenizer ports tokenizers.uk.UkrainianSRXSentenceTokenizer.
type UkrainianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewUkrainianSRXSentenceTokenizer() *UkrainianSRXSentenceTokenizer {
	return &UkrainianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("uk"),
	}
}
