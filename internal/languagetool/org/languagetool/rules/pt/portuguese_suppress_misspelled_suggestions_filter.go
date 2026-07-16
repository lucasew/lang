package pt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// PortugueseSuppressMisspelledSuggestionsFilter wraps the core suppress-misspelled filter.
type PortugueseSuppressMisspelledSuggestionsFilter struct {
	*rules.SuppressMisspelledSuggestionsFilter
}

func NewPortugueseSuppressMisspelledSuggestionsFilter() *PortugueseSuppressMisspelledSuggestionsFilter {
	return &PortugueseSuppressMisspelledSuggestionsFilter{
		SuppressMisspelledSuggestionsFilter: rules.NewSuppressMisspelledSuggestionsFilter(),
	}
}
