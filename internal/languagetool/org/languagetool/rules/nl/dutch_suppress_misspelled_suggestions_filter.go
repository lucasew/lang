package nl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// DutchSuppressMisspelledSuggestionsFilter wraps the core suppress-misspelled filter.
type DutchSuppressMisspelledSuggestionsFilter struct {
	*rules.SuppressMisspelledSuggestionsFilter
}

func NewDutchSuppressMisspelledSuggestionsFilter() *DutchSuppressMisspelledSuggestionsFilter {
	return &DutchSuppressMisspelledSuggestionsFilter{
		SuppressMisspelledSuggestionsFilter: rules.NewSuppressMisspelledSuggestionsFilter(),
	}
}
