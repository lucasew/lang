package en

// Twin of EnglishWordRepeatBeginningRuleTest — PRP inject for pronoun suggestions (Java).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func engWRBMessages() map[string]string {
	return map[string]string{
		"desc_repetition_beginning_adv":       "Three successive sentences begin with the same adverb.",
		"desc_repetition_beginning_word":      "Three successive sentences begin with the same word.",
		"desc_repetition_beginning_thesaurus": "Consider using a thesaurus to find synonyms.",
	}
}

func TestEnglishWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewEnglishWordRepeatBeginningRule(engWRBMessages())

	matches := rule.MatchList(languagetool.SplitAndAnalyze("This is good. This is good, too."))
	require.Equal(t, 0, len(matches))
	matches = rule.MatchList(languagetool.SplitAndAnalyze("The car. The bicycle. The third sentence with 'the'."))
	require.Equal(t, 0, len(matches))

	// "I" / "He" need PRP POS for pronoun suggestions (Java hasPosTag("PRP"))
	matches = rule.MatchList(analyzeENWRB("I think so. I have seen that before. I don't like it."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Furthermore, I", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Likewise, I", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "Not only that, but I", matches[0].GetSuggestedReplacements()[2])

	matches = rule.MatchList(analyzeENWRB("He thinks so. He has seen that before. He doesn't like it."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Furthermore, he", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Likewise, he", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "Not only that, but he", matches[0].GetSuggestedReplacements()[2])

	// without PRP: still may match as word-repeat, but no pronoun-style suggestions
	matches = rule.MatchList(languagetool.SplitAndAnalyze("I think so. I have seen that before. I don't like it."))
	require.Equal(t, 1, len(matches))
	require.Empty(t, matches[0].GetSuggestedReplacements())

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

// Java EnglishWordRepeatBeginningRule: id + Moreover example pair.
func TestEnglishWordRepeatBeginningRule_Metadata(t *testing.T) {
	rule := NewEnglishWordRepeatBeginningRule(nil)
	require.Equal(t, "ENGLISH_WORD_REPEAT_BEGINNING_RULE", rule.GetID())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Moreover, the street is almost entirely residential. <marker>Moreover</marker>, it was named after a poet.", inc[0].GetExample())
	require.Equal(t, []string{"It"}, inc[0].GetCorrections())
	require.Equal(t, "Moreover, the street is almost entirely residential. <marker>It</marker> was named after a poet.", rule.GetCorrectExamples()[0].GetExample())
}

// analyzeENWRB injects PRP on first content token of each sentence when surface is a personal pronoun form.
func analyzeENWRB(text string) []*languagetool.AnalyzedSentence {
	parts := languagetool.SplitAndAnalyze(text)
	for _, s := range parts {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok == nil || tok.IsSentenceStart() {
				continue
			}
			// first content word
			t := tok.GetToken()
			switch t {
			case "I", "You", "He", "She", "It", "We", "They",
				"Me", "Him", "Her", "Us", "Them":
				pos := "PRP"
				tok.AddReading(languagetool.NewAnalyzedToken(t, &pos, nil), "test")
			}
			_ = strings.ToLower
			break
		}
	}
	return parts
}
