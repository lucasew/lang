package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/GenericUnpairedBracketsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGenericUnpairedBracketsRule_DutchRule(t *testing.T) {
	rule := NewDutchUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	// correct
	require.Equal(t, 0, matchN(" Eurlings: “De gegevens van de dienst zijn van cruciaal belang voor de veiligheid van de luchtvaart en de scheepvaart”."))
	require.Equal(t, 0, matchN(" Eurlings: \u201eDe gegevens van de dienst zijn van cruciaal belang voor de veiligheid van de luchtvaart en de scheepvaart\u201d."))
	// incorrect
	require.Equal(t, 1, matchN("Het centrale probleem van het werk is de „dichterlijke kuischheid."))
	require.Equal(t, 1, matchN(" Eurlings: “De gegevens van de dienst zijn van cruciaal belang voor de veiligheid van de luchtvaart en de scheepvaart."))
}
