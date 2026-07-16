package de

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// GermanSRXSentenceTokenizer ports tokenizers.de.GermanSRXSentenceTokenizer.
type GermanSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewGermanSRXSentenceTokenizer() *GermanSRXSentenceTokenizer {
	return &GermanSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("de"),
	}
}
