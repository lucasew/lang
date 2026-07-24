package fr

// Twin of languagetool-language-modules/fr/src/test/java/org/languagetool/rules/fr/QuestionWhitespaceStrictRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestQuestionWhitespaceStrictRule_Rule(t *testing.T) {
	rule := NewQuestionWhitespaceStrictRule(nil)

	// correct (nbsp/fine):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("C'est vrai\u00a0!"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Qu'est ce que c'est\u00a0?"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("L'enjeu de ce livre est donc triple\u00a0: philosophique"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Bonjour :)"))))

	// covered by non-strict (missing space entirely → no match for strict):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("C'est vrai!"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Qu'est ce que c'est?"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("L'enjeu de ce livre est donc triple: philosophique;"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Le guillemet ouvrant est suivi d'un espace insécable\u00a0: «mais le lieu [...] et le guillemet fermant est précédé d'un espace insécable\u00a0: [...] littérature»."))))

	// regular space before ; (strict should flag):
	matches := rule.Match(languagetool.AnalyzePlain("L'enjeu de ce livre est donc triple: philosophique ;"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 50, matches[0].GetFromPos())
	require.Equal(t, 52, matches[0].GetToPos())
	require.Equal(t, "\u202f;", matches[0].GetSuggestedReplacements()[0])

	// regular space before ! ?
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("C'est vrai !"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Qu'est ce que c'est ?"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Bonjour : )"))))

	// guillemets with regular spaces
	matches = rule.Match(languagetool.AnalyzePlain("Le guillemet ouvrant est suivi d'un espace insécable\u00a0: « mais le lieu [...] et le guillemet fermant est précédé d'un espace insécable\u00a0: [...] littérature »."))
	require.Equal(t, 2, len(matches))
}
