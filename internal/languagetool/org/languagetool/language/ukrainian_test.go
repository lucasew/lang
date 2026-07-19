package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUkrainianLanguage(t *testing.T) {
	require.Equal(t, "uk", UkrainianLanguageDefault.GetShortCode())
	require.Equal(t, "Ukrainian", UkrainianLanguageDefault.GetName())
	require.Contains(t, UkrainianLanguageDefault.GetCountries(), "UA")
	require.True(t, UkrainianIgnoredChars.MatchString("\u0301"))
	require.Contains(t, UkrainianLanguageDefault.RuleFiles, "grammar-spelling.xml")
	// GetRuleFileNames: grammar.xml + RULE_FILES in Java order
	files := UkrainianLanguageDefault.GetRuleFileNames()
	require.Equal(t, []string{
		"/org/languagetool/rules/uk/grammar.xml",
		"/org/languagetool/rules/uk/grammar-spelling.xml",
		"/org/languagetool/rules/uk/grammar-grammar.xml",
		"/org/languagetool/rules/uk/grammar-barbarism.xml",
		"/org/languagetool/rules/uk/grammar-style.xml",
		"/org/languagetool/rules/uk/grammar-punctuation.xml",
	}, files)
}
