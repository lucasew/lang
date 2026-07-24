package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildShortCodeWithCountryAndVariant(t *testing.T) {
	// Java Language.buildShortCodeWithCountryAndVariant
	require.Equal(t, "en-US", BuildShortCodeWithCountryAndVariant("en", []string{"US"}, ""))
	require.Equal(t, "ca-ES-valencia", BuildShortCodeWithCountryAndVariant("ca", []string{"ES"}, "valencia"))
	require.Equal(t, "ca-ES-balear", BuildShortCodeWithCountryAndVariant("ca", []string{"ES"}, "balear"))
	require.Equal(t, "eo", BuildShortCodeWithCountryAndVariant("eo", nil, ""))
	require.Equal(t, "eo", BuildShortCodeWithCountryAndVariant("eo", []string{}, ""))
	// private-use tag unchanged
	require.Equal(t, "de-DE-x-simple-language",
		BuildShortCodeWithCountryAndVariant("de-DE-x-simple-language", []string{"DE"}, ""))
	// multi-country: no append (Java requires length == 1)
	require.Equal(t, "xx", BuildShortCodeWithCountryAndVariant("xx", []string{"A", "B"}, ""))
}

func TestShortCodeWithCountryAndVariant_Languages(t *testing.T) {
	require.Equal(t, "en", AmericanEnglish.GetShortCode())
	require.Equal(t, "en-US", AmericanEnglish.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "en/common_words.txt", AmericanEnglish.GetCommonWordsPath())

	require.Equal(t, "de", GermanyGerman.GetShortCode())
	require.Equal(t, "de-DE", GermanyGerman.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "de/common_words.txt", GermanyGerman.GetCommonWordsPath())

	require.Equal(t, "fr", FrenchFrance.GetShortCode())
	require.Equal(t, "fr-FR", FrenchFrance.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "fr-CA", CanadianFrench.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "fr-BE", BelgianFrench.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "fr/common_words.txt", FrenchFrance.GetCommonWordsPath())

	require.Equal(t, "es", SpanishSpain.GetShortCode())
	require.Equal(t, "es-ES", SpanishSpain.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "es-AR", SpanishVoseo.GetShortCodeWithCountryAndVariant())

	require.Equal(t, "pt", PortugalPortuguese.GetShortCode())
	require.Equal(t, "pt-PT", PortugalPortuguese.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "pt-BR", BrazilianPortuguese.GetShortCodeWithCountryAndVariant())

	require.Equal(t, "nl", DutchNetherlands.GetShortCode())
	require.Equal(t, "nl-NL", DutchNetherlands.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "nl", BelgianDutch.GetShortCode())
	require.Equal(t, "nl-BE", BelgianDutch.GetShortCodeWithCountryAndVariant())

	require.Equal(t, "ca", Catalan.GetShortCode())
	require.Equal(t, "ca-ES", Catalan.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "ca-ES-valencia", ValencianCatalan.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "ca-ES-balear", BalearicCatalan.GetShortCodeWithCountryAndVariant())

	require.Equal(t, "it-IT", Italian.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "pl-PL", Polish.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "ru-RU", Russian.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "uk-UA", UkrainianLanguageDefault.GetShortCodeWithCountryAndVariant())
	// Java Serbian base: empty countries → "sr"; SerbianSerbian → "sr-RS"
	require.Equal(t, "sr", DefaultSerbian.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "sr-RS", SerbianSerbia.GetShortCodeWithCountryAndVariant())

	// SmallLang
	require.Equal(t, "sk-SK", Slovak.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "be-BY", Belarusian.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "eo", Esperanto.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "sk/common_words.txt", Slovak.GetCommonWordsPath())
	// Java Khmer/Japanese return null
	require.Equal(t, CommonWordsPathNone, Khmer.GetCommonWordsPath())
	require.Equal(t, CommonWordsPathNone, Japanese.GetCommonWordsPath())
}
