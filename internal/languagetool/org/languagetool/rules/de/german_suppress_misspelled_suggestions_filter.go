package de

// GermanSuppressMisspelledSuggestionsFilter ports the DE subclass of
// AbstractSuppressMisspelledSuggestionsFilter without tagger/speller:
// keeps all suggestions when IsMisspelled is nil.
type GermanSuppressMisspelledSuggestionsFilter struct {
	IsMisspelled func(word string) bool
}

func NewGermanSuppressMisspelledSuggestionsFilter() *GermanSuppressMisspelledSuggestionsFilter {
	return &GermanSuppressMisspelledSuggestionsFilter{}
}

// FilterSuggestions drops misspelled suggestions; empty result means suppress match
// when suppressMatch is true (default).
func (f *GermanSuppressMisspelledSuggestionsFilter) FilterSuggestions(suggs []string, suppressMatch bool) (kept []string, keepMatch bool) {
	miss := f.IsMisspelled
	if miss == nil {
		// without dict keep all suggestions
		return suggs, true
	}
	for _, s := range suggs {
		if !miss(s) {
			kept = append(kept, s)
		}
	}
	if len(kept) == 0 && suppressMatch {
		return nil, false
	}
	return kept, true
}
