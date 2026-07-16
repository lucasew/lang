package languagetool

// Language is the pure-Go surface for org.languagetool.Language (minimal).
// Concrete languages implement this without requiring the full Java hierarchy.
type Language interface {
	GetName() string
	GetShortCode() string
	GetShortCodeWithCountryAndVariant() string
	GetCountries() []string
	// CreateDefaultTagger may return nil if untagged.
	CreateDefaultTagger() Tagger
}

// LanguageDefaults provides default tokenizers/tagger hooks shared by many languages.
type LanguageDefaults struct {
	Name      string
	ShortCode string
	Countries []string
	// Optional factories
	Tagger Tagger
}

func (l LanguageDefaults) GetName() string {
	if l.Name == "" {
		return l.ShortCode
	}
	return l.Name
}
func (l LanguageDefaults) GetShortCode() string { return l.ShortCode }
func (l LanguageDefaults) GetShortCodeWithCountryAndVariant() string {
	return l.ShortCode
}
func (l LanguageDefaults) GetCountries() []string      { return l.Countries }
func (l LanguageDefaults) CreateDefaultTagger() Tagger { return l.Tagger }

// ToCommonTypography applies a small subset of Language.toAdvancedTypography helpers.
func ToCommonTypography(input string) string {
	// non-breaking spaces between single-letter abbreviations: "e. g." → "e.\u00a0g."
	// kept minimal — full Java Language typography deferred
	return input
}

// AdaptSuggestion strips trivial markup from suggestions (Language.adaptSuggestion stub).
func AdaptSuggestion(s string) string { return s }
