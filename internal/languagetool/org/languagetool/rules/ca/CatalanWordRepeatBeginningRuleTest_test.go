package ca

// Unit coverage for CatalanWordRepeatBeginningRule (no dedicated Java twin beyond example).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewCatalanWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_adv":       "Tres frases successives comencen amb el mateix adverbi.",
		"desc_repetition_beginning_word":      "Tres frases successives comencen amb la mateixa paraula.",
		"desc_repetition_beginning_thesaurus": "Considereu un diccionari de sinònims.",
	})
	// article exception
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("El cotxe. El bus. El tren."))))
	// contrast adverb pair
	matches := rule.MatchList(languagetool.SplitAndAnalyze("Però el carrer és modernista. Però té nom de poeta."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "Així i tot")
	// three jo
	matches = rule.MatchList(languagetool.SplitAndAnalyze("Jo penso. Jo veig. Jo actuo."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "A més a més, jo", matches[0].GetSuggestedReplacements()[0])
}
