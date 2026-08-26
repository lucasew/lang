package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/CheckCaseRuleTest.java
// Surface ASR2 CheckingCase port — some mid-sentence / punctuation edge cases may differ from Java.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCheckCaseRule_Rule(t *testing.T) {
	rule := NewCheckCaseRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Sap que tinc dos bons amics?"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("El país necessita tecnologia més moderna."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Da Vinci"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Amb Joan Pau i Josep Maria."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("ESTAT D'ALARMA"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Educació Secundària Obligatòria"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Educació secundària obligatòria"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("d'educació secundària obligatòria"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Els drets humans"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Declaració Universal dels Drets Humans"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("El codi Da Vinci"))))

	// partial wrong casing of multiword proper phrase
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Educació Secundària obligatòria"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Declaració Universal dels drets humans"))))

	matches := rule.Match(languagetool.AnalyzePlain("Joan pau"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Joan Pau", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Expedient de Regulació Temporal d'Ocupació"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Expedient de regulació temporal d'ocupació", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Em vaig entrevistar amb Joan maria"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Joan Maria", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Em vaig entrevistar amb Da Vinci"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "da Vinci", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Baixar Al-Assad"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Baixar al-Assad", matches[0].GetSuggestedReplacements()[0])
}
