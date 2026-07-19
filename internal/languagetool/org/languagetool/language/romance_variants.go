package language

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// FrenchVariant ports French locale languages.
type FrenchVariant struct {
	ShortCode     string
	Name          string
	Countries     []string
	SpellerRuleID string
}

func (v FrenchVariant) GetName() string { return v.Name }

// GetShortCode ports French.getShortCode ("fr" for all locales).
func (v FrenchVariant) GetShortCode() string { return "fr" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v FrenchVariant) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("fr", v.Countries, "")
}

func (v FrenchVariant) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → fr/common_words.txt.
func (v FrenchVariant) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

var (
	// SpellerRuleID ports MorfologikFrenchSpellerRule.getId → FR_SPELLING_RULE.
	FrenchFrance   = FrenchVariant{ShortCode: "fr", Name: "French", Countries: []string{"FR"}, SpellerRuleID: "FR_SPELLING_RULE"}
	CanadianFrench = FrenchVariant{ShortCode: "fr-CA", Name: "French (Canada)", Countries: []string{"CA"}, SpellerRuleID: "FR_SPELLING_RULE"}
	SwissFrench    = FrenchVariant{ShortCode: "fr-CH", Name: "French (Switzerland)", Countries: []string{"CH"}, SpellerRuleID: "FR_SPELLING_RULE"}
	BelgianFrench  = FrenchVariant{ShortCode: "fr-BE", Name: "French (Belgium)", Countries: []string{"BE"}, SpellerRuleID: "FR_SPELLING_RULE"}
)

// GetDefaultSpellingRuleID ports MorfologikFrenchSpellerRule getId.
func (v FrenchVariant) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "FR_SPELLING_RULE"
}

func AllFrenchVariants() []FrenchVariant {
	return []FrenchVariant{FrenchFrance, CanadianFrench, SwissFrench, BelgianFrench}
}

// SpanishVariant ports Spanish locales.
type SpanishVariant struct {
	ShortCode     string
	Name          string
	Countries     []string
	SpellerRuleID string
	Voseo         bool
}

func (v SpanishVariant) GetName() string { return v.Name }

// GetShortCode ports Spanish.getShortCode ("es" for all locales).
func (v SpanishVariant) GetShortCode() string { return "es" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v SpanishVariant) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("es", v.Countries, "")
}

func (v SpanishVariant) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → es/common_words.txt.
func (v SpanishVariant) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

var (
	SpanishSpain = SpanishVariant{ShortCode: "es", Name: "Spanish", Countries: []string{"ES"}, SpellerRuleID: "MORFOLOGIK_RULE_ES"}
	SpanishVoseo = SpanishVariant{ShortCode: "es-AR", Name: "Spanish (voseo)", Countries: []string{"AR"}, SpellerRuleID: "MORFOLOGIK_RULE_ES", Voseo: true}
)

// GetDefaultSpellingRuleID ports MorfologikSpanishSpellerRule getId.
func (v SpanishVariant) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_ES"
}

// GetOpeningDoubleQuote ports Spanish.getOpeningDoubleQuote ("«").
func (v SpanishVariant) GetOpeningDoubleQuote() string { return "«" }

// GetClosingDoubleQuote ports Spanish.getClosingDoubleQuote ("»").
func (v SpanishVariant) GetClosingDoubleQuote() string { return "»" }

// GetOpeningSingleQuote ports Spanish.getOpeningSingleQuote ("‘").
func (v SpanishVariant) GetOpeningSingleQuote() string { return "‘" }

// GetClosingSingleQuote ports Spanish.getClosingSingleQuote ("’").
func (v SpanishVariant) GetClosingSingleQuote() string { return "’" }

// IsAdvancedTypographyEnabled ports Spanish.isAdvancedTypographyEnabled (true).
func (v SpanishVariant) IsAdvancedTypographyEnabled() bool { return true }

// ToAdvancedTypography ports Spanish base Language.toAdvancedTypography with Spanish quotes.
func (v SpanishVariant) ToAdvancedTypography(input string) string {
	return SpanishAdvancedTypography(input)
}

// HasMinMatchesRules ports Spanish.hasMinMatchesRules (true).
func (v SpanishVariant) HasMinMatchesRules() bool { return true }

func AllSpanishVariants() []SpanishVariant {
	return []SpanishVariant{SpanishSpain, SpanishVoseo}
}

// PortugueseVariant ports Portuguese locales.
type PortugueseVariant struct {
	ShortCode     string
	Name          string
	Countries     []string
	SpellerRuleID string
}

func (v PortugueseVariant) GetName() string { return v.Name }

// GetShortCode ports Portuguese.getShortCode ("pt" for all locales).
func (v PortugueseVariant) GetShortCode() string { return "pt" }

// GetShortCodeWithCountryAndVariant: stored ShortCode is full locale (pt-PT, pt-BR, …).
func (v PortugueseVariant) GetShortCodeWithCountryAndVariant() string {
	if v.ShortCode != "" {
		return v.ShortCode
	}
	return BuildShortCodeWithCountryAndVariant("pt", v.Countries, "")
}

func (v PortugueseVariant) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → pt/common_words.txt.
func (v PortugueseVariant) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports locale MorfologikPortuguese*SpellerRule getId.
func (v PortugueseVariant) GetDefaultSpellingRuleID() string {
	return v.SpellerRuleID
}

// GetOpeningDoubleQuote ports Portuguese / PortugalPortuguese.
// Base Portuguese: “ ; PortugalPortuguese overrides to «.
func (v PortugueseVariant) GetOpeningDoubleQuote() string {
	if isPortugalPortugueseCode(v.ShortCode) {
		return "«"
	}
	return "“"
}

// GetClosingDoubleQuote ports Portuguese / PortugalPortuguese (” vs »).
func (v PortugueseVariant) GetClosingDoubleQuote() string {
	if isPortugalPortugueseCode(v.ShortCode) {
		return "»"
	}
	return "”"
}

// GetOpeningSingleQuote ports Portuguese.getOpeningSingleQuote ("‘").
func (v PortugueseVariant) GetOpeningSingleQuote() string { return "‘" }

// GetClosingSingleQuote ports Portuguese.getClosingSingleQuote ("’").
func (v PortugueseVariant) GetClosingSingleQuote() string { return "’" }

// IsAdvancedTypographyEnabled ports Portuguese.isAdvancedTypographyEnabled (true).
func (v PortugueseVariant) IsAdvancedTypographyEnabled() bool { return true }

// ToAdvancedTypography ports Portuguese.toAdvancedTypography with locale quotes.
func (v PortugueseVariant) ToAdvancedTypography(input string) string {
	if isPortugalPortugueseCode(v.ShortCode) {
		return PortugalPortugueseAdvancedTypography(input)
	}
	return PortugueseAdvancedTypography(input)
}

var (
	PortugalPortuguese   = PortugueseVariant{ShortCode: "pt-PT", Name: "Portuguese (Portugal)", Countries: []string{"PT"}, SpellerRuleID: "MORFOLOGIK_RULE_PT_PT"}
	BrazilianPortuguese  = PortugueseVariant{ShortCode: "pt-BR", Name: "Portuguese (Brazil)", Countries: []string{"BR"}, SpellerRuleID: "MORFOLOGIK_RULE_PT_BR"}
	AngolaPortuguese     = PortugueseVariant{ShortCode: "pt-AO", Name: "Portuguese (Angola)", Countries: []string{"AO"}, SpellerRuleID: "MORFOLOGIK_RULE_PT_AO"}
	MozambiquePortuguese = PortugueseVariant{ShortCode: "pt-MZ", Name: "Portuguese (Mozambique)", Countries: []string{"MZ"}, SpellerRuleID: "MORFOLOGIK_RULE_PT_MZ"}
)

func AllPortugueseVariants() []PortugueseVariant {
	return []PortugueseVariant{PortugalPortuguese, BrazilianPortuguese, AngolaPortuguese, MozambiquePortuguese}
}

// GetMaintainedState ports French.getMaintainedState → ActivelyMaintained.
func (v FrenchVariant) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports French.getMaintainers → Dominique Pellé.
func (v FrenchVariant) GetMaintainers() []Contributor {
	return []Contributor{DominiquePelle}
}

// GetMaintainedState ports Spanish.getMaintainedState → ActivelyMaintained.
func (v SpanishVariant) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Spanish.getMaintainers → Jaume Ortolà.
func (v SpanishVariant) GetMaintainers() []Contributor {
	return []Contributor{NewContributor("Jaume Ortolà")}
}

// GetMaintainedState ports Portuguese.getMaintainedState → ActivelyMaintained.
func (v PortugueseVariant) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Portuguese.getMaintainers.
func (v PortugueseVariant) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributorWithURL("Marco A.G. Pinto", "http://www.marcoagpinto.com/"),
		NewContributor("Susana Boatto (pt-BR)"),
		NewContributorWithURL("Tiago F. Santos (3.6-4.7)", "https://github.com/TiagoSantos81"),
		NewContributorWithURL("Matheus Poletto (pt-BR)", "https://github.com/MatheusPoletto"),
	}
}
