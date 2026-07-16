package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// EnglishSuppressMisspelledSuggestionsFilter wraps the core suppress-misspelled filter.
type EnglishSuppressMisspelledSuggestionsFilter struct {
	*rules.SuppressMisspelledSuggestionsFilter
}

func NewEnglishSuppressMisspelledSuggestionsFilter() *EnglishSuppressMisspelledSuggestionsFilter {
	return &EnglishSuppressMisspelledSuggestionsFilter{
		SuppressMisspelledSuggestionsFilter: rules.NewSuppressMisspelledSuggestionsFilter(),
	}
}
