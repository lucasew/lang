package pt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// PortugueseSRXSentenceTokenizer ports tokenizers.pt.PortugueseSRXSentenceTokenizer.
type PortugueseSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewPortugueseSRXSentenceTokenizer() *PortugueseSRXSentenceTokenizer {
	return &PortugueseSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("pt"),
	}
}
