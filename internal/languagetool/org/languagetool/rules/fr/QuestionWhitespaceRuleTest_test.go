package fr

// Twin of languagetool-language-modules/fr/src/test/java/org/languagetool/rules/fr/QuestionWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestQuestionWhitespaceRule_Rule(t *testing.T) {
	rule := NewQuestionWhitespaceRule(nil)

	// correct sentences (nbsp / fine space or accepted whitespace):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("C'est vrai\u00a0!"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Qu'est ce que c'est\u00a0?"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("L'enjeu de ce livre est donc triple\u00a0: philosophique"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Bonjour :)"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("5/08/2019 23:30"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("C'est vrai\u00a0!!"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("C'est vrai\u00a0??"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("☀️9:00"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("00:80:41:ae:fd:7e"))))

	// Also accept regular space before ?! for non-strict rule:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("C'est vrai !"))))

	// errors:
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("C'est vrai!"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Qu'est ce que c'est?"))))
	require.Equal(t, 2, len(rule.Match(languagetool.AnalyzePlain("L'enjeu de ce livre est donc triple: philosophique;"))))

	matches1 := rule.Match(languagetool.AnalyzePlain("L'enjeu de ce livre est donc triple: philosophique ;"))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 29, matches1[0].GetFromPos())
	require.Equal(t, 36, matches1[0].GetToPos())
	require.Equal(t, "triple\u00a0:", matches1[0].GetSuggestedReplacements()[0])

	// guillemets:
	matches2 := rule.Match(languagetool.AnalyzePlain("LanguageTool offre une «vérification» orthographique."))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 23, matches2[0].GetFromPos())
	require.Equal(t, 37, matches2[0].GetToPos())
	require.Equal(t, "«\u00a0vérification\u00a0»", matches2[0].GetSuggestedReplacements()[0])

	matches3 := rule.Match(languagetool.AnalyzePlain("Le guillemet ouvrant est suivi d'un espace insécable\u00a0: «mais le lieu [...] et le guillemet fermant est précédé d'un espace insécable\u00a0: [...] littérature»."))
	require.Equal(t, 2, len(matches3))
	require.Equal(t, "«\u00a0mais", matches3[0].GetSuggestedReplacements()[0])
	require.Equal(t, 55, matches3[0].GetFromPos())
	require.Equal(t, 60, matches3[0].GetToPos())
	require.Equal(t, "littérature\u00a0»", matches3[1].GetSuggestedReplacements()[0])
	require.Equal(t, 141, matches3[1].GetFromPos())
	require.Equal(t, 153, matches3[1].GetToPos())
}
