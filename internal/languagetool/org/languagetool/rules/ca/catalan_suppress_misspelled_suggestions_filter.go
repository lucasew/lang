package ca

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// CatalanSuppressMisspelledSuggestionsFilter wraps the core suppress-misspelled filter.
type CatalanSuppressMisspelledSuggestionsFilter struct {
	*rules.SuppressMisspelledSuggestionsFilter
}

func NewCatalanSuppressMisspelledSuggestionsFilter() *CatalanSuppressMisspelledSuggestionsFilter {
	return &CatalanSuppressMisspelledSuggestionsFilter{
		SuppressMisspelledSuggestionsFilter: rules.NewSuppressMisspelledSuggestionsFilter(),
	}
}
