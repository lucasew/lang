package sk

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// SlovakSRXSentenceTokenizer ports tokenizers.sk.SlovakSRXSentenceTokenizer.
type SlovakSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewSlovakSRXSentenceTokenizer() *SlovakSRXSentenceTokenizer {
	return &SlovakSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("sk"),
	}
}
