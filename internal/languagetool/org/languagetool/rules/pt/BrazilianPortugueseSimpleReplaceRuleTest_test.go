package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/BrazilianPortugueseSimpleReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestBrazilianPortugueseSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewBrazilianPortugueseReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Fui de ônibus até o açougue italiano."))))

	check := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "sentence %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0])
	}
	check("Vou de autocarro.", "ônibus")
	check("O lançamento de dardo é um desporto.", "esporte")
	check("Está no meu ADN!", "DNA")

	// named-entity exceptions (Java uses NP tags)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("José António Miranda Coutinho"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Jerónimo Soares"))))
}
