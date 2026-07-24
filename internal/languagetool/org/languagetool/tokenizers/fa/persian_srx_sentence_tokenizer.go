package fa

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// PersianSRXSentenceTokenizer ports tokenizers.fa.PersianSRXSentenceTokenizer.
type PersianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewPersianSRXSentenceTokenizer() *PersianSRXSentenceTokenizer {
	return &PersianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("fa"),
	}
}
