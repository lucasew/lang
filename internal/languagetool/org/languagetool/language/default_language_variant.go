package language

// Default language variant short codes from Language.getDefaultLanguageVariant().
// Values are Languages short codes (getShortCodeWithCountryAndVariant style) as used
// in Java Languages.getLanguageForShortCode / *English.getInstance() paths.

// EnglishDefaultLanguageVariantCode ports English.getDefaultLanguageVariant → AmericanEnglish.
func EnglishDefaultLanguageVariantCode() string { return "en-US" }

// GermanDefaultLanguageVariantCode ports German.getDefaultLanguageVariant → GermanyGerman.
func GermanDefaultLanguageVariantCode() string { return "de-DE" }

// FrenchDefaultLanguageVariantCode ports French.getDefaultLanguageVariant → "fr".
func FrenchDefaultLanguageVariantCode() string { return "fr" }

// SpanishDefaultLanguageVariantCode ports Spanish.getDefaultLanguageVariant → "es".
func SpanishDefaultLanguageVariantCode() string { return "es" }

// PortugueseDefaultLanguageVariantCode ports Portuguese.getDefaultLanguageVariant → "pt-PT".
func PortugueseDefaultLanguageVariantCode() string { return "pt-PT" }

// DutchDefaultLanguageVariantCode ports Dutch.getDefaultLanguageVariant → "nl".
func DutchDefaultLanguageVariantCode() string { return "nl" }

// CatalanDefaultLanguageVariantCode ports Catalan.getDefaultLanguageVariant → "ca-ES".
func CatalanDefaultLanguageVariantCode() string { return "ca-ES" }

// SerbianDefaultLanguageVariantCode ports Serbian.getDefaultLanguageVariant → SerbianSerbian.
// SerbianSerbian keeps shortCode "sr" with countries RS → shortCodeWithCountryAndVariant "sr-RS".
func SerbianDefaultLanguageVariantCode() string { return "sr-RS" }

// IrishDefaultLanguageVariantCode ports Irish.getDefaultLanguageVariant → getInstance() (self).
func IrishDefaultLanguageVariantCode() string { return "ga" }

// GetDefaultLanguageVariantCode on EnglishVariant returns the language-family default (en-US),
// not the receiver's own code (matches English base method, not per-locale overrides).
func (v EnglishVariant) GetDefaultLanguageVariantCode() string {
	return EnglishDefaultLanguageVariantCode()
}

// GetDefaultLanguageVariantCode ports German.getDefaultLanguageVariant → de-DE.
func (v GermanVariant) GetDefaultLanguageVariantCode() string {
	return GermanDefaultLanguageVariantCode()
}

// GetDefaultLanguageVariantCode ports French.getDefaultLanguageVariant → fr.
func (v FrenchVariant) GetDefaultLanguageVariantCode() string {
	return FrenchDefaultLanguageVariantCode()
}

// GetDefaultLanguageVariantCode ports Spanish.getDefaultLanguageVariant → es.
func (v SpanishVariant) GetDefaultLanguageVariantCode() string {
	return SpanishDefaultLanguageVariantCode()
}

// GetDefaultLanguageVariantCode ports Portuguese.getDefaultLanguageVariant → pt-PT.
func (v PortugueseVariant) GetDefaultLanguageVariantCode() string {
	return PortugueseDefaultLanguageVariantCode()
}

// GetDefaultLanguageVariantCode ports Dutch.getDefaultLanguageVariant → nl.
func (v DutchVariant) GetDefaultLanguageVariantCode() string {
	return DutchDefaultLanguageVariantCode()
}

// GetDefaultLanguageVariantCode ports Catalan.getDefaultLanguageVariant → ca-ES.
func (v CatalanVariant) GetDefaultLanguageVariantCode() string {
	return CatalanDefaultLanguageVariantCode()
}

// GetDefaultLanguageVariantCode ports Serbian.getDefaultLanguageVariant → sr-RS.
func (s Serbian) GetDefaultLanguageVariantCode() string {
	return SerbianDefaultLanguageVariantCode()
}
