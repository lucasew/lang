package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/UkrainianTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

// Port of UkrainianTest.testLanguage — metadata surface without full JLanguageTool.
func TestUkrainian_Language(t *testing.T) {
	uk := language.UkrainianLanguageDefault
	require.Equal(t, "uk", uk.GetShortCode())
	require.Equal(t, "Ukrainian", uk.GetName())
	require.Contains(t, uk.GetCountries(), "UA")
	require.Equal(t, "MORFOLOGIK_RULE_UK_UA", uk.SpellerRuleID)
	require.NotEmpty(t, uk.RuleFiles)
	require.Contains(t, uk.RuleFiles, "grammar-grammar.xml")
	// ignored chars include soft hyphen / combining acute
	require.True(t, language.UkrainianIgnoredChars.MatchString("\u00AD"))
	require.True(t, language.UkrainianIgnoredChars.MatchString("\u0301"))
}
