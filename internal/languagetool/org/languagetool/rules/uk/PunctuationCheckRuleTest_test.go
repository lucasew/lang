package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/PunctuationCheckRuleTest.java
//
// Java:
//   PunctuationCheckRule rule = new PunctuationCheckRule(TestTools.getEnglishMessages());
//   JLanguageTool lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   matches = rule.match(lt.getAnalyzedSentence(...));
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of PunctuationCheckRuleTest.testRule.
func TestPunctuationCheckRule_Rule(t *testing.T) {
	// Java: new PunctuationCheckRule(TestTools.getEnglishMessages()) — messages only for category.
	rule := NewPunctuationCheckRule(nil)
	// Java: new JLanguageTool(Ukrainian.DEFAULT_VARIANT)
	lt := languagetool.NewJLanguageTool("uk")

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Дві, коми. Ось: дві!!!"))))

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("- Це ваша пряма мова?!!"))))

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Дві,- коми!.."))))

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Таке питання?.."))))

	// correct sentences:
	// Java: "Два  пробіли." // поки що ігноруємо - не царська це справа :)
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Два  пробіли."))))

	// incorrect sentences:
	matches := rule.Match(lt.GetAnalyzedSentence("Дві крапки.."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, ".", matches[0].GetSuggestedReplacements()[0])

	// incorrect sentences:
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("Дві,, коми."))))

	// incorrect sentences:
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("Не там ,кома."))))

	// incorrect sentences:
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("Двокрапка:- з тире."))))
}
