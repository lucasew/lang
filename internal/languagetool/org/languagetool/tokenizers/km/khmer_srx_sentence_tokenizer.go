package km

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// KhmerSRXSentenceTokenizer ports tokenizers.km.KhmerSRXSentenceTokenizer.
type KhmerSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewKhmerSRXSentenceTokenizer() *KhmerSRXSentenceTokenizer {
	return &KhmerSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("km"),
	}
}
