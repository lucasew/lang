package ru

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// RussianSuppressMisspelledSuggestionsFilter wraps the core suppress-misspelled filter.
type RussianSuppressMisspelledSuggestionsFilter struct {
	*rules.SuppressMisspelledSuggestionsFilter
}

func NewRussianSuppressMisspelledSuggestionsFilter() *RussianSuppressMisspelledSuggestionsFilter {
	return &RussianSuppressMisspelledSuggestionsFilter{
		SuppressMisspelledSuggestionsFilter: rules.NewSuppressMisspelledSuggestionsFilter(),
	}
}
