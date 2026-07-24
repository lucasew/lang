package language

// Twin of SerbianTest.getRuleFileNames
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerbian_GetRuleFileNames(t *testing.T) {
	want := []string{
		"/org/languagetool/rules/sr/grammar.xml",
		"/org/languagetool/rules/sr/grammar-barbarism.xml",
		"/org/languagetool/rules/sr/grammar-logical.xml",
		"/org/languagetool/rules/sr/grammar-punctuation.xml",
		"/org/languagetool/rules/sr/grammar-spelling.xml",
		"/org/languagetool/rules/sr/grammar-style.xml",
	}
	require.Equal(t, want, NewSerbian().GetRuleFileNames())
	require.Equal(t, "sr", DefaultSerbian.GetShortCode())
	require.Equal(t, "Serbian", DefaultSerbian.GetName())
	// Java Serbian.getCountries → empty; shortCodeWithCountryAndVariant → "sr"
	require.Empty(t, DefaultSerbian.GetCountries())
	require.Equal(t, "sr", DefaultSerbian.GetShortCodeWithCountryAndVariant())
}

func TestSerbianVariants_Metadata(t *testing.T) {
	// SerbianSerbian default country variant
	require.Equal(t, "Serbian (Serbia)", SerbianSerbia.GetName())
	require.Equal(t, []string{"RS"}, SerbianSerbia.GetCountries())
	require.Equal(t, "sr-RS", SerbianSerbia.GetShortCodeWithCountryAndVariant())
	require.False(t, SerbianSerbia.Jekavian)
	require.Equal(t, "MORFOLOGIK_RULE_SR_EKAVIAN", SerbianSerbia.GetDefaultSpellingRuleID())
	require.Contains(t, SerbianSerbia.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_SR_EKAVIAN")
	require.Contains(t, SerbianSerbia.GetRelevantRuleIDs(), "SR_EKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE")

	// Jekavian dialect
	require.True(t, JekavianSerbian.Jekavian)
	require.Equal(t, "sr", JekavianSerbian.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "MORFOLOGIK_RULE_SR_JEKAVIAN", JekavianSerbian.GetDefaultSpellingRuleID())
	jek := JekavianSerbian.GetRelevantRuleIDs()
	require.Contains(t, jek, "MORFOLOGIK_RULE_SR_JEKAVIAN")
	require.Contains(t, jek, "SR_JEKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE")
	require.Contains(t, jek, "SR_JEKAVIAN_SIMPLE_STYLE_REPLACE_RULE")
	require.NotContains(t, jek, "MORFOLOGIK_RULE_SR_EKAVIAN")

	// Country variants of Jekavian
	require.Equal(t, "Serbian (Bosnia and Herzegovina)", BosnianSerbian.GetName())
	require.Equal(t, "sr-BA", BosnianSerbian.GetShortCodeWithCountryAndVariant())
	require.True(t, BosnianSerbian.Jekavian)
	require.Equal(t, "sr-HR", CroatianSerbian.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "sr-ME", MontenegrinSerbian.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "Serbian (Montenegro)", MontenegrinSerbian.GetName())

	require.Len(t, AllSerbianVariants(), 6)
	// default language variant is SerbianSerbian (sr-RS)
	require.Equal(t, "sr-RS", DefaultSerbian.GetDefaultLanguageVariantCode())
}
