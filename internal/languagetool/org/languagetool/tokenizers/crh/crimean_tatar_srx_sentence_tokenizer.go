package crh

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// CrimeanTatarSRXSentenceTokenizer ports tokenizers.crh.CrimeanTatarSRXSentenceTokenizer.
type CrimeanTatarSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewCrimeanTatarSRXSentenceTokenizer() *CrimeanTatarSRXSentenceTokenizer {
	return &CrimeanTatarSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("crh"),
	}
}
