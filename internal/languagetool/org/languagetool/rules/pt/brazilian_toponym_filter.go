package pt

// BrazilianToponymFilter ports org.languagetool.rules.pt.BrazilianToponymFilter.
type BrazilianToponymFilter struct {
	Map *BrazilianToponymMap
}

func NewBrazilianToponymFilter() *BrazilianToponymFilter {
	return &BrazilianToponymFilter{Map: LoadBrazilianToponymMap()}
}

// Suggest returns the en-dash + state suggestion when the toponym is valid
// and the underlined text is not already that suggestion.
// Empty string means suppress the match.
func (f *BrazilianToponymFilter) Suggest(toponym, underlined, state string) string {
	suggestion := "–" + state
	if suggestion == underlined {
		return ""
	}
	if f.Map == nil || !f.Map.IsValidToponym(toponym) {
		return ""
	}
	return suggestion
}
