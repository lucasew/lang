package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/ContractionSpellingRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestContractionSpellingRule_Rule(t *testing.T) {
	rule := NewContractionSpellingRule(nil)

	// correct
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("It wasn't me."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("I'm ill."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Staatszerfall im südlichen Afrika."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("by IVE"))))
	// Java disambiguation IGNORE_WHYS_AND_WHEREFORES → ignore_spelling on "whys"
	// (not surface invent TokenException).
	require.Equal(t, 0, len(rule.Match(withIgnoreSpelling("Never mind the whys and wherefores.", "whys"))))

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

func TestContractionSpellingRule_FailClosedWithoutDisambig(t *testing.T) {
	rule := NewContractionSpellingRule(nil)
	// Without ignore_spelling from disambiguation, "whys" is in contractions.txt.
	matches := rule.Match(languagetool.AnalyzePlain("Never mind the whys and wherefores."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "why's", matches[0].GetSuggestedReplacements()[0])
}

// Java ContractionSpellingRule ctor: TYPOS, Misspelling, URL, example pair.
func TestContractionSpellingRule_Metadata(t *testing.T) {
	rule := NewContractionSpellingRule(nil)
	require.Equal(t, "Spelling of English contractions", rule.GetDescription())
	require.Contains(t, rule.GetURL(), "grammar-contractions")
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "We <marker>havent</marker> earned anything.", inc[0].GetExample())
	require.Equal(t, []string{"haven't"}, inc[0].GetCorrections())
	require.Equal(t, "We <marker>haven't</marker> earned anything.", rule.GetCorrectExamples()[0].GetExample())
}

func withIgnoreSpelling(text, surface string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == surface {
			tok.IgnoreSpelling()
		}
	}
	return sent
}
