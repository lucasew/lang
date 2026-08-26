package language

import (
	"regexp"
	"strings"
)

// caOldDiacritics ports Catalan.CA_OLD_DIACRITICS (CASE_INSENSITIVE | UNICODE_CASE).
// Matches suggestions that contain traditional orthography forms to normalize.
var caOldDiacritics = regexp.MustCompile(`(?i).*\b(sóc|dóna|dónes|vénen|véns|fóra|adéu|féu|desféu|vés|contrapèl)\b.*`)

// CatalanRemoveOldDiacritics ports Catalan.removeOldDiacritics (IEC orthography).
// Java is king — exact replace pairs only, no invent.
func CatalanRemoveOldDiacritics(s string) string {
	repl := []struct{ old, new string }{
		{"contrapèl", "contrapel"},
		{"Contrapèl", "Contrapel"},
		{"vés", "ves"},
		{"féu", "feu"},
		{"desféu", "desfeu"},
		{"adéu", "adeu"},
		{"dóna", "dona"},
		{"dónes", "dones"},
		{"sóc", "soc"},
		{"vénen", "venen"},
		// Java Catalan.removeOldDiacritics: .replace("véns", "véns") — no-op (same string).
		// Do not invent véns→vens; only the capitalized Véns→Vens pair is effective.
		{"véns", "véns"},
		{"fóra", "fora"},
		{"Vés", "Ves"},
		{"Féu", "Feu"},
		{"Desféu", "Desfeu"},
		{"Adéu", "Adeu"},
		{"Dóna", "Dona"},
		{"Dónes", "Dones"},
		{"Sóc", "Soc"},
		{"Vénen", "Venen"},
		{"Véns", "Vens"},
		{"Fóra", "Fora"},
	}
	for _, p := range repl {
		s = strings.ReplaceAll(s, p.old, p.new)
	}
	return s
}

// CatalanSuggestionNeedsOldDiacriticStrip reports CA_OLD_DIACRITICS.matches(s).
func CatalanSuggestionNeedsOldDiacriticStrip(s string) bool {
	return caOldDiacritics.MatchString(s)
}
