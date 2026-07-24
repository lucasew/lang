package be

// Twin of languagetool-language-modules/be/src/test/java/org/languagetool/rules/be/BelarusianSpecificCaseRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestBelarusianSpecificCaseRule_Rule(t *testing.T) {
	rule := NewBelarusianSpecificCaseRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Беларуская Народная Рэспубліка"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Папа Рымскі"))))

	// Full lowercase of multiword proper names
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("вярхоўны суд рэспублікі беларусь"))))

	matches := rule.Match(languagetool.AnalyzePlain("Мне падабаецца air France."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 15, matches[0].GetFromPos())
	require.Equal(t, 25, matches[0].GetToPos())
	require.Equal(t, "Air France", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Уласныя імёны і назвы пішуцца з вялікай літары.", matches[0].GetMessage())
}
