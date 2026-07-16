package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseClicheRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseClicheRule_Rule(t *testing.T) {
	rule := NewPortugueseClicheRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Evite as frases-feitas e as expressões idiomáticas."))))

	check := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q matches %d", sentence, len(matches))
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0])
	}
	check("Teste. A todo o vapor!", "O mais rápido possível")
	check("A todo o vapor!", "O mais rápido possível")
	check("Teste, a todo o vapor!", "o mais rápido possível")
}
