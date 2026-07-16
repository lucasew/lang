package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UppercaseNounReadingFilter ports UppercaseNounReadingFilter without a tagger.
// Accepts when the uppercased token looks noun-like (capital letter word).
// Optional HasNounReading overrides tagger behavior when set.
type UppercaseNounReadingFilter struct {
	HasNounReading func(uppercased string) bool
}

func NewUppercaseNounReadingFilter() *UppercaseNounReadingFilter {
	return &UppercaseNounReadingFilter{}
}

// Accept returns true if the match should be kept for token.
func (f *UppercaseNounReadingFilter) Accept(token string) bool {
	if token == "" {
		panic("token required for UppercaseNounReadingFilter")
	}
	upper := tools.UppercaseFirstChar(token)
	if f.HasNounReading != nil {
		return f.HasNounReading(upper)
	}
	// soft: accept any alphabetic capitalizable form (without tagger)
	return upper != "" && tools.StartsWithUppercase(upper)
}
