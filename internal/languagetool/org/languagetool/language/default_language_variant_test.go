package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultLanguageVariantCodes(t *testing.T) {
	// Java getDefaultLanguageVariant short codes
	require.Equal(t, "en-US", EnglishDefaultLanguageVariantCode())
	require.Equal(t, "de-DE", GermanDefaultLanguageVariantCode())
	require.Equal(t, "fr", FrenchDefaultLanguageVariantCode())
	require.Equal(t, "es", SpanishDefaultLanguageVariantCode())
	require.Equal(t, "pt-PT", PortugueseDefaultLanguageVariantCode())
	require.Equal(t, "nl", DutchDefaultLanguageVariantCode())
	require.Equal(t, "ca-ES", CatalanDefaultLanguageVariantCode())
	require.Equal(t, "sr-RS", SerbianDefaultLanguageVariantCode())
	require.Equal(t, "ga", IrishDefaultLanguageVariantCode())

	// Methods on types (family default, not self)
	require.Equal(t, "en-US", BritishEnglish.GetDefaultLanguageVariantCode())
	require.Equal(t, "en-US", AmericanEnglish.GetDefaultLanguageVariantCode())
	require.Equal(t, "de-DE", SwissGerman.GetDefaultLanguageVariantCode())
	require.Equal(t, "de-DE", GermanyGerman.GetDefaultLanguageVariantCode())
	require.Equal(t, "fr", BelgianFrench.GetDefaultLanguageVariantCode())
	require.Equal(t, "es", SpanishVoseo.GetDefaultLanguageVariantCode())
	require.Equal(t, "pt-PT", BrazilianPortuguese.GetDefaultLanguageVariantCode())
	require.Equal(t, "nl", BelgianDutch.GetDefaultLanguageVariantCode())
	require.Equal(t, "ca-ES", ValencianCatalan.GetDefaultLanguageVariantCode())
	require.Equal(t, "sr-RS", DefaultSerbian.GetDefaultLanguageVariantCode())
}
