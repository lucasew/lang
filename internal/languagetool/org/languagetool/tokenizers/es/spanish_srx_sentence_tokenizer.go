package es

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// SpanishSRXSentenceTokenizer ports tokenizers.es.SpanishSRXSentenceTokenizer.
type SpanishSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewSpanishSRXSentenceTokenizer() *SpanishSRXSentenceTokenizer {
	return &SpanishSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("es"),
	}
}
