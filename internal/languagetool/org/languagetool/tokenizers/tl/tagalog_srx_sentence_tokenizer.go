package tl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// TagalogSRXSentenceTokenizer ports tokenizers.tl.TagalogSRXSentenceTokenizer.
type TagalogSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewTagalogSRXSentenceTokenizer() *TagalogSRXSentenceTokenizer {
	return &TagalogSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("tl"),
	}
}
