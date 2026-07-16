package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/SimpleReplaceRuleTest.java
// Without gender/number filter; assertions use surface replacements from replace.txt.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Això està força bé."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Joan Navarro no és de Navarra ni de Jerez."))))

	matches := rule.Match(languagetool.AnalyzePlain("El recader fa huelga."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, "ordinari", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "transportista", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "vaga", matches[1].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Aconteixements"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Esdeveniments", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Els desencontres."))
	require.Equal(t, 1, len(matches))
	// order from file; Java gender filter may reorder articles — check set membership
	require.Contains(t, matches[0].GetSuggestedReplacements(), "desavinences")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "desacords")

	matches = rule.Match(languagetool.AnalyzePlain("La seguent solució."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "següent")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "seient")

	matches = rule.Match(languagetool.AnalyzePlain("Un caminet poc ciclable baixa uns metres."))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "pedalable")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "ciclista")

	matches = rule.Match(languagetool.AnalyzePlain("La seva escola transformada pq les seves filles encaixen molt bé."))
	require.Equal(t, "perquè", matches[0].GetSuggestedReplacements()[0])
}
