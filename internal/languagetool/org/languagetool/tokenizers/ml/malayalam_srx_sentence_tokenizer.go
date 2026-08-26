package ml

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// MalayalamSRXSentenceTokenizer ports tokenizers.ml.MalayalamSRXSentenceTokenizer.
type MalayalamSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewMalayalamSRXSentenceTokenizer() *MalayalamSRXSentenceTokenizer {
	return &MalayalamSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("ml"),
	}
}
