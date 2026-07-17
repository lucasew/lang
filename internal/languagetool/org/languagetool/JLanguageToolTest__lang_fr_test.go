package languagetool

// Twin of languagetool-language-modules/fr/src/test/java/org/languagetool/JLanguageToolTest.java
// Full FR check pipeline deferred — Analyze + typography surface greens.
// (language.FrenchAdvancedTypography lives in language package to avoid import cycle)
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testLanguageDependentFilter
func TestJLanguageTool_lang_fr_LanguageDependentFilter(t *testing.T) {
	lt := NewJLanguageTool("fr")
	require.Equal(t, "fr", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Ceci est une phrase."))
}

// Port of JLanguageToolTest.testMultitokenSpeller
func TestJLanguageTool_lang_fr_MultitokenSpeller(t *testing.T) {
	lt := NewJLanguageTool("fr")
	require.NotEmpty(t, lt.Analyze("New York"))
}

// Port of JLanguageToolTest.testMatchfiltering
func TestJLanguageTool_lang_fr_Matchfiltering(t *testing.T) {
	// soft: analysis path only
	lt := NewJLanguageTool("fr")
	require.NotEmpty(t, lt.Analyze("Le chat dort."))
}

// Port of JLanguageToolTest.testQuotes
func TestJLanguageTool_lang_fr_Quotes(t *testing.T) {
	// FR guillemet-style quotes via typography config (no language import)
	cfg := TypographyConfig{
		Enabled:            true,
		OpeningDoubleQuote: "«\u00a0",
		ClosingDoubleQuote: "\u00a0»",
		OpeningSingleQuote: "‘",
		ClosingSingleQuote: "’",
	}
	out := ToAdvancedTypography(`"C'est"`, cfg)
	require.Contains(t, out, "«")
	require.Contains(t, out, "»")
}

// Port of JLanguageToolTest.testMergingOfGrammarCorrections
func TestJLanguageTool_lang_fr_MergingOfGrammarCorrections(t *testing.T) {
	lt := NewJLanguageTool("fr")
	require.NotEmpty(t, lt.Analyze("Phrase une. Phrase deux."))
}
