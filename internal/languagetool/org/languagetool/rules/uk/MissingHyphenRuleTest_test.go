package uk

// Twin of MissingHyphenRuleTest — surface prefix path (no word tagger gate).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMissingHyphenRule_Rule(t *testing.T) {
	rule := NewMissingHyphenRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Поїхали у штаб квартиру."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "штаб-квартиру", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Роблю тайм аут"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "тайм-аут", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("Такий компакт диск."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "компакт-диск", matches[0].GetSuggestedReplacements()[0])

	// :alt → join without hyphen
	matches = rule.Match(languagetool.AnalyzePlain("Такий міні автомобіль."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "мініавтомобіль", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("всі медіа півострова."))))
}
