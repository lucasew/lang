package en

// AdverbFilter ports org.languagetool.rules.en.AdverbFilter.
// Maps adverb+noun pattern args to adjective + noun suggestion.
type AdverbFilter struct{}

func NewAdverbFilter() *AdverbFilter {
	return &AdverbFilter{}
}

// Suggest returns "adjective noun" when the adverb maps and differs from the adjective.
// Empty string means keep match without changing suggestion (Java leaves match as-is).
func (f *AdverbFilter) Suggest(adverb, noun string) string {
	adj, ok := adverb2Adj[adverb]
	if !ok || adj == adverb {
		return ""
	}
	return adj + " " + noun
}
