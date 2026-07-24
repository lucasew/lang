package ca

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/language"

// AdaptSuggestion ports Catalan.adaptSuggestion (Language implementation).
// Single source of truth: language.CatalanAdaptSuggestion.
func AdaptSuggestion(s, originalErrorStr string) string {
	return language.CatalanAdaptSuggestion(s, originalErrorStr)
}
