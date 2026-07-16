package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/PunctuationCheckRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPunctuationCheckRule_Rule(t *testing.T) {
	rule := NewPunctuationCheckRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Дві, коми. Ось: дві!!!"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("- Це ваша пряма мова?!!"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Дві,- коми!.."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Таке питання?.."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Два  пробіли."))))

	matches := rule.Match(languagetool.AnalyzePlain("Дві крапки.."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, ".", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Дві,, коми."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Не там ,кома."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Двокрапка:- з тире."))))
}
