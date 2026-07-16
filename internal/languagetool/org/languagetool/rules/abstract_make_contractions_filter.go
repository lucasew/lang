package rules

// MakeContractionsFilter ports AbstractMakeContractionsFilter: rewrites suggestions
// via a language-specific FixContractions function.
type MakeContractionsFilter struct {
	FixContractions func(suggestion string) string
}

func NewMakeContractionsFilter(fix func(string) string) *MakeContractionsFilter {
	if fix == nil {
		fix = func(s string) string { return s }
	}
	return &MakeContractionsFilter{FixContractions: fix}
}

// MapSuggestions applies FixContractions to each suggestion.
func (f *MakeContractionsFilter) MapSuggestions(suggs []string) []string {
	out := make([]string, len(suggs))
	for i, s := range suggs {
		out[i] = f.FixContractions(s)
	}
	return out
}
