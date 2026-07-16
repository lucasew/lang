package fr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// FrenchSuppressMisspelledSuggestionsFilter wraps the core suppress-misspelled filter.
type FrenchSuppressMisspelledSuggestionsFilter struct {
	*rules.SuppressMisspelledSuggestionsFilter
}

func NewFrenchSuppressMisspelledSuggestionsFilter() *FrenchSuppressMisspelledSuggestionsFilter {
	return &FrenchSuppressMisspelledSuggestionsFilter{
		SuppressMisspelledSuggestionsFilter: rules.NewSuppressMisspelledSuggestionsFilter(),
	}
}
