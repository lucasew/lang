package es

// Twin of languagetool-language-modules/es/src/test/java/org/languagetool/rules/es/SpanishWrongWordInContextRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpanishWrongWordInContextRule_Rule(t *testing.T) {
	rule := NewSpanishWrongWordInContextRule(nil)

	matches := rule.Match(languagetool.AnalyzePlain("Le infringió un duro castigo"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "infligió", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Infligía todas las normas."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Infringía", matches[0].GetSuggestedReplacements()[0])

	// baca / vaca
	matches = rule.Match(languagetool.AnalyzePlain("La baca da leche."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, "vaca", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Pon la maleta en la vaca."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "baca", matches[0].GetSuggestedReplacements()[0])
}
