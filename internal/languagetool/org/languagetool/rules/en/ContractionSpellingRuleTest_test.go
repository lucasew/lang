package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/ContractionSpellingRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestContractionSpellingRule_Rule(t *testing.T) {
	rule := NewContractionSpellingRule(nil)

	// correct
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("It wasn't me."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("I'm ill."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Staatszerfall im südlichen Afrika."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("by IVE"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Never mind the whys and wherefores."))))

	checkSimple := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "matches for %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()), "repl count for %q", sentence)
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0], "replacement for %q", sentence)
	}

	checkSimple("Wasnt this great", "Wasn't")
	checkSimple("YOURE WRONG", "YOU'RE")
	checkSimple("Dont do this", "Don't")
	checkSimple("It wasnt me", "wasn't")
	checkSimple("You neednt do this", "needn't")
	checkSimple("I know Im wrong", "I'm")

	matches := rule.Match(languagetool.AnalyzePlain("Whereve you are"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 2, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, "Where've", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Wherever", matches[0].GetSuggestedReplacements()[1])
}
