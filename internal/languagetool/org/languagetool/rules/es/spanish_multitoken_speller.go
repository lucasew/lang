package es

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"

// SpanishMultitokenSpeller ports org.languagetool.rules.es.SpanishMultitokenSpeller.
type SpanishMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

func NewSpanishMultitokenSpeller() *SpanishMultitokenSpeller {
	return &SpanishMultitokenSpeller{MultitokenSpeller: multitoken.NewMultitokenSpeller()}
}
