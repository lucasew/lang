package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWordRepeatBeginningRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func engWRBMessages() map[string]string {
	return map[string]string{
		"desc_repetition_beginning_adv":      "Three successive sentences begin with the same adverb.",
		"desc_repetition_beginning_word":     "Three successive sentences begin with the same word.",
		"desc_repetition_beginning_thesaurus": "Consider using a thesaurus to find synonyms.",
	}
}

func TestEnglishWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewEnglishWordRepeatBeginningRule(engWRBMessages())

	// two successive sentences that start with the same non-adverb word.
	matches := rule.MatchList(languagetool.SplitAndAnalyze("This is good. This is good, too."))
	require.Equal(t, 0, len(matches))
	// three successive sentences that start with the same exception word ("the").
	matches = rule.MatchList(languagetool.SplitAndAnalyze("The car. The bicycle. The third sentence with 'the'."))
	require.Equal(t, 0, len(matches))

	// three successive sentences that start with personal pronoun "I"
	matches = rule.MatchList(languagetool.SplitAndAnalyze("I think so. I have seen that before. I don't like it."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Furthermore, I", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Likewise, I", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "Not only that, but I", matches[0].GetSuggestedReplacements()[2])

	// three successive with "He"
	matches = rule.MatchList(languagetool.SplitAndAnalyze("He thinks so. He has seen that before. He doesn't like it."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Furthermore, he", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Likewise, he", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "Not only that, but he", matches[0].GetSuggestedReplacements()[2])

	// two successive sentences that start with adverb "Also"
	matches = rule.MatchList(languagetool.SplitAndAnalyze("Also, I play football. Also, I play basketball."))
	require.Equal(t, 1, len(matches))
	suggs := matches[0].GetSuggestedReplacements()
	has := func(s string) bool {
		for _, x := range suggs {
			if x == s {
				return true
			}
		}
		return false
	}
	require.True(t, has("Additionally"))
	require.True(t, has("Besides"))
	require.True(t, has("Furthermore"))
	require.True(t, has("Moreover"))
	require.True(t, has("In addition"))
	require.True(t, has("As well as"))
}
