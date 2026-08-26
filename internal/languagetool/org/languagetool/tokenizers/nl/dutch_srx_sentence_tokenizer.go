package nl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// DutchSRXSentenceTokenizer ports tokenizers.nl.DutchSRXSentenceTokenizer.
type DutchSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewDutchSRXSentenceTokenizer() *DutchSRXSentenceTokenizer {
	return &DutchSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("nl"),
	}
}
