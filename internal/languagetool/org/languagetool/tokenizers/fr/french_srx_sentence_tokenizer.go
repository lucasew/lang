package fr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// FrenchSRXSentenceTokenizer ports tokenizers.fr.FrenchSRXSentenceTokenizer.
type FrenchSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewFrenchSRXSentenceTokenizer() *FrenchSRXSentenceTokenizer {
	return &FrenchSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("fr"),
	}
}
