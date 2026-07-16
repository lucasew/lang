package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/SimpleReplaceBalearicRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceBalearicRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceBalearicRule(nil)

	// correct sentences
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Això està força bé."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Joan Navarro no és de Navarra ni de Jerez."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Prosper Mérimée."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Index Librorum Prohibitorum"))))

	// incorrect sentences
	matches := rule.Match(languagetool.AnalyzePlain("El calcul del telefon."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, "càlcul", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "telèfon", matches[1].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("EL CALCUL DEL TELEFON."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, "CÀLCUL", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "TELÈFON", matches[1].GetSuggestedReplacements()[0])
}
