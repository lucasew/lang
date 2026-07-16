package es

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// SpanishSuppressMisspelledSuggestionsFilter wraps the core suppress-misspelled filter.
type SpanishSuppressMisspelledSuggestionsFilter struct {
	*rules.SuppressMisspelledSuggestionsFilter
}

func NewSpanishSuppressMisspelledSuggestionsFilter() *SpanishSuppressMisspelledSuggestionsFilter {
	return &SpanishSuppressMisspelledSuggestionsFilter{
		SuppressMisspelledSuggestionsFilter: rules.NewSuppressMisspelledSuggestionsFilter(),
	}
}
