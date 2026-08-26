package ru

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// RussianSRXSentenceTokenizer ports tokenizers.ru.RussianSRXSentenceTokenizer.
type RussianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewRussianSRXSentenceTokenizer() *RussianSRXSentenceTokenizer {
	return &RussianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("ru"),
	}
}
