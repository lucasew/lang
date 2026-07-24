package ca

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// CatalanSRXSentenceTokenizer ports tokenizers.ca.CatalanSRXSentenceTokenizer.
type CatalanSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewCatalanSRXSentenceTokenizer() *CatalanSRXSentenceTokenizer {
	return &CatalanSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("ca"),
	}
}
