package el

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// GreekSRXSentenceTokenizer ports tokenizers.el.GreekSRXSentenceTokenizer.
type GreekSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewGreekSRXSentenceTokenizer() *GreekSRXSentenceTokenizer {
	return &GreekSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("el"),
	}
}
