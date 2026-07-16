package rules

// AdaptSuggestionsFilter ports org.languagetool.rules.AdaptSuggestionsFilter.
// Adapt maps (replacement, originalError) → adjusted suggestion.
type AdaptSuggestionsFilter struct {
	Adapt func(replacement, originalError string) string
}

func NewAdaptSuggestionsFilter(adapt func(string, string) string) *AdaptSuggestionsFilter {
	if adapt == nil {
		adapt = func(s, _ string) string { return s }
	}
	return &AdaptSuggestionsFilter{Adapt: adapt}
}

// MapSuggestions rewrites each suggestion using Adapt.
func (f *AdaptSuggestionsFilter) MapSuggestions(suggs []string, originalError string) []string {
	out := make([]string, len(suggs))
	for i, s := range suggs {
		out[i] = f.Adapt(s, originalError)
	}
	return out
}
