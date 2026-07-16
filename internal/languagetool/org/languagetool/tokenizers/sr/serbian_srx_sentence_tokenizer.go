package sr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// SerbianSRXSentenceTokenizer ports tokenizers.sr.SerbianSRXSentenceTokenizer.
type SerbianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewSerbianSRXSentenceTokenizer() *SerbianSRXSentenceTokenizer {
	return &SerbianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("sr"),
	}
}
