package language

// Twin of languagetool-standalone LanguageTest — metadata surface.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguage_GetShortNameWithVariant(t *testing.T) {
	require.Equal(t, "en-US", AmericanEnglish.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "en-GB", BritishEnglish.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "de-DE", GermanyGerman.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "de-AT", AustrianGerman.GetShortCodeWithCountryAndVariant())
}

func TestLanguage_Equals(t *testing.T) {
	require.Equal(t, AmericanEnglish.ShortCode, NewAmericanEnglish().ShortCode)
	require.NotEqual(t, AmericanEnglish.ShortCode, BritishEnglish.ShortCode)
}

func TestLanguage_EqualsConsiderVariantIfSpecified(t *testing.T) {
	a, ok := EnglishVariantByCode("en-US")
	require.True(t, ok)
	b, ok := EnglishVariantByCode("en-us")
	require.True(t, ok)
	require.Equal(t, a.ShortCode, b.ShortCode)
	_, ok = EnglishVariantByCode("en-GB")
	require.True(t, ok)
	require.NotEqual(t, AmericanEnglish.ShortCode, BritishEnglish.ShortCode)
}

func TestLanguage_RuleFileName(t *testing.T) {
	require.NotEmpty(t, AmericanEnglish.Name)
	require.Contains(t, AmericanEnglish.SpellerRuleID, "EN_US")
}

func TestLanguage_GetTranslatedName(t *testing.T) {
	require.Equal(t, "English (US)", AmericanEnglish.GetName())
	require.Equal(t, "English (British)", BritishEnglish.GetName())
}

func TestLanguage_CreateDefaultJLanguageTool(t *testing.T) {
	t.Skip("unimplemented: full JLanguageTool factory for language variants")
}
