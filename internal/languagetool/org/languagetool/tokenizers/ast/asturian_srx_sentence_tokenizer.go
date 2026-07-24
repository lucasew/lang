package ast

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// AsturianSRXSentenceTokenizer ports tokenizers.ast.AsturianSRXSentenceTokenizer.
type AsturianSRXSentenceTokenizer struct {
	*tokenizers.SRXSentenceTokenizer
}

func NewAsturianSRXSentenceTokenizer() *AsturianSRXSentenceTokenizer {
	return &AsturianSRXSentenceTokenizer{
		SRXSentenceTokenizer: tokenizers.NewSRXSentenceTokenizer("ast"),
	}
}
