package da

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// DanishSRXSentenceTokenizer ports tokenizers.da.DanishSRXSentenceTokenizer.
type DanishSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewDanishSRXSentenceTokenizer() *DanishSRXSentenceTokenizer {
	return &DanishSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("da"),
	}
}
