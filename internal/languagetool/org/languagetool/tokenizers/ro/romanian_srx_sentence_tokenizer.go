package ro

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// RomanianSRXSentenceTokenizer ports tokenizers.ro.RomanianSRXSentenceTokenizer.
type RomanianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewRomanianSRXSentenceTokenizer() *RomanianSRXSentenceTokenizer {
	return &RomanianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("ro"),
	}
}
