package sv

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// SwedishSRXSentenceTokenizer ports tokenizers.sv.SwedishSRXSentenceTokenizer.
type SwedishSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewSwedishSRXSentenceTokenizer() *SwedishSRXSentenceTokenizer {
	return &SwedishSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("sv"),
	}
}
