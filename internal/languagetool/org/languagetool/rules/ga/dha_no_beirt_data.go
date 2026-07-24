package ga

// DhaNoBeirtData ports org.languagetool.rules.ga.DhaNoBeirtData.
type DhaNoBeirtData struct{}

// GetDaoine returns human nouns used with personal numbers (from people.txt).
func (DhaNoBeirtData) GetDaoine() map[string]bool {
	return loadPeople()
}

// GetNumberReplacements returns cardinal → personal number forms.
func (DhaNoBeirtData) GetNumberReplacements() map[string]string {
	// copy so callers cannot mutate package state
	out := make(map[string]string, len(numberReplacements))
	for k, v := range numberReplacements {
		out[k] = v
	}
	return out
}
