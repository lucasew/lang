package ca

// Unit tests for Catalan SimpleReplaceMultiwordsRule (no dedicated Java twin test).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceMultiwordsRule(t *testing.T) {
	rule := NewSimpleReplaceMultiwordsRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Aire condicionat a l'oficina."))))

	matches := rule.Match(languagetool.AnalyzePlain("Van comprar aire a condicionat."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "aire condicionat", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("És el modus operandis habitual."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "modus operandi", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Hidrats carboni a la dieta."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Hidrats de carboni", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Menja hidrats carboni."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "hidrats de carboni", matches[0].GetSuggestedReplacements()[0])
}
