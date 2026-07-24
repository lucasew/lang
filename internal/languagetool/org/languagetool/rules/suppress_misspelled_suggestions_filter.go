package rules

// SuppressMisspelledSuggestionsFilter ports AbstractSuppressMisspelledSuggestionsFilter
// without tagger/postag filtering: drops misspelled suggestions; may suppress match.
type SuppressMisspelledSuggestionsFilter struct {
	// IsMisspelled reports unknown words; nil keeps all suggestions.
	IsMisspelled func(word string) bool
}

func NewSuppressMisspelledSuggestionsFilter() *SuppressMisspelledSuggestionsFilter {
	return &SuppressMisspelledSuggestionsFilter{}
}

// FilterSuggestions returns kept suggestions and whether the rule match should remain.
// suppressMatch (default true) drops the whole match when no suggestions remain.
func (f *SuppressMisspelledSuggestionsFilter) FilterSuggestions(suggs []string, suppressMatch bool) (kept []string, keepMatch bool) {
	miss := f.IsMisspelled
	if miss == nil {
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
