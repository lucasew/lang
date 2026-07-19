package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpanishAdvancedTypography(t *testing.T) {
	require.True(t, SpanishIsAdvancedTypographyEnabled())
	require.Equal(t, "…", SpanishAdvancedTypography("..."))
	require.Equal(t, "Di «hola»", SpanishAdvancedTypography(`Di "hola"`))
	require.True(t, SpanishHasMinMatchesRules())
}

func TestDutchAdvancedTypography(t *testing.T) {
	require.True(t, DutchIsAdvancedTypographyEnabled())
	require.Equal(t, "Zeg “hallo”", DutchAdvancedTypography(`Zeg "hallo"`))
	require.Equal(t, "…", DutchAdvancedTypography("..."))
	require.Equal(t, "“", DutchNetherlands.GetOpeningDoubleQuote())
	require.Equal(t, "Zeg “hallo”", DutchNetherlands.ToAdvancedTypography(`Zeg "hallo"`))
	require.True(t, DutchNetherlands.IsAdvancedTypographyEnabled())
}

func TestPortugueseAdvancedTypography(t *testing.T) {
	require.True(t, PortugueseIsAdvancedTypographyEnabled())
	// Base Portuguese (e.g. pt-BR): curly doubles
	require.Equal(t, "Diz “olá”", PortugueseAdvancedTypography(`Diz "olá"`))
	require.Equal(t, "“", BrazilianPortuguese.GetOpeningDoubleQuote())
	require.Equal(t, "Diz “olá”", BrazilianPortuguese.ToAdvancedTypography(`Diz "olá"`))
	// PortugalPortuguese overrides doubles to guillemets « »
	require.Equal(t, "«", PortugalPortuguese.GetOpeningDoubleQuote())
	require.Equal(t, "»", PortugalPortuguese.GetClosingDoubleQuote())
	require.Equal(t, "‘", PortugalPortuguese.GetOpeningSingleQuote())
	require.True(t, PortugalPortuguese.IsAdvancedTypographyEnabled())
	require.Equal(t, "Diz «olá»", PortugalPortuguese.ToAdvancedTypography(`Diz "olá"`))
	require.Equal(t, "Diz «olá»", PortugalPortugueseAdvancedTypography(`Diz "olá"`))
	// Default Portuguese alias is PortugalPortuguese
	require.Equal(t, "«", Portuguese.GetOpeningDoubleQuote())
}

func TestRussianAdvancedTypography(t *testing.T) {
	require.True(t, RussianIsAdvancedTypographyEnabled())
	require.Equal(t, "Скажи «привет»", RussianAdvancedTypography(`Скажи "привет"`))
	require.Equal(t, "«", Russian.GetOpeningDoubleQuote())
	require.Equal(t, "Скажи «привет»", Russian.ToAdvancedTypography(`Скажи "привет"`))
	require.True(t, Russian.IsAdvancedTypographyEnabled())
}

func TestBelarusianAdvancedTypography(t *testing.T) {
	require.True(t, BelarusianIsAdvancedTypographyEnabled())
	require.Equal(t, "Скажы «прывітанне»", BelarusianAdvancedTypography(`Скажы "прывітанне"`))
	require.Equal(t, "…", BelarusianAdvancedTypography("..."))
	require.Equal(t, "«", Belarusian.GetOpeningDoubleQuote())
	require.True(t, Belarusian.IsAdvancedTypographyEnabled())
	require.Equal(t, "Скажы «прывітанне»", Belarusian.ToAdvancedTypography(`Скажы "прывітанне"`))
}

func TestSpanishVariantTypography(t *testing.T) {
	require.Equal(t, "«", SpanishSpain.GetOpeningDoubleQuote())
	require.True(t, SpanishSpain.IsAdvancedTypographyEnabled())
	require.True(t, SpanishSpain.HasMinMatchesRules())
	require.Equal(t, "Di «hola»", SpanishSpain.ToAdvancedTypography(`Di "hola"`))
}

func TestUkrainianLanguageTypographyQuotes(t *testing.T) {
	// Quotes defined but isAdvancedTypographyEnabled=false
	require.Equal(t, "«", UkrainianLanguageDefault.GetOpeningDoubleQuote())
	require.False(t, UkrainianLanguageDefault.IsAdvancedTypographyEnabled())
	require.Equal(t, "«ok»", UkrainianLanguageDefault.ToAdvancedTypography("<suggestion>ok</suggestion>"))
	require.Equal(t, "A...", UkrainianLanguageDefault.ToAdvancedTypography("A..."))
}

func TestEnglishAdvancedTypography(t *testing.T) {
	require.True(t, EnglishIsAdvancedTypographyEnabled())
	require.Equal(t, "Say “hello”", EnglishAdvancedTypography(`Say "hello"`))
	require.Equal(t, "…", EnglishAdvancedTypography("..."))
	require.True(t, EnglishHasMinMatchesRules())
	// EnglishVariant surface (AmericanEnglish)
	require.Equal(t, "“", AmericanEnglish.GetOpeningDoubleQuote())
	require.True(t, AmericanEnglish.IsAdvancedTypographyEnabled())
	require.True(t, AmericanEnglish.HasMinMatchesRules())
	require.Equal(t, -50, AmericanEnglish.GetDefaultRulePriorityForStyle())
	require.Equal(t, "Say “hello”", AmericanEnglish.ToAdvancedTypography(`Say "hello"`))
}

func TestCatalanHasMinMatchesRules(t *testing.T) {
	require.True(t, CatalanHasMinMatchesRules())
	require.Equal(t, "«", Catalan.GetOpeningDoubleQuote())
	require.True(t, Catalan.IsAdvancedTypographyEnabled())
	require.True(t, Catalan.HasMinMatchesRules())
	require.Equal(t, -50, Catalan.GetDefaultRulePriorityForStyle())
	require.Equal(t, "Digues «hola»", Catalan.ToAdvancedTypography(`Digues "hola"`))
}

func TestFrenchVariantTypographyQuotes(t *testing.T) {
	require.Equal(t, "«", FrenchFrance.GetOpeningDoubleQuote())
	require.Equal(t, "‘", FrenchFrance.GetOpeningSingleQuote())
	require.True(t, FrenchFrance.IsAdvancedTypographyEnabled())
	require.True(t, FrenchFrance.HasMinMatchesRules())
}
