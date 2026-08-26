package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/ContractionSpellingRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// TestContractionSpellingRule_Rule ports ContractionSpellingRuleTest.testRule (all active cases).
func TestContractionSpellingRule_Rule(t *testing.T) {
	rule := NewContractionSpellingRule(nil)

	// correct sentences (Java assertEquals(0, …))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("It wasn't me."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("I'm ill."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Staatszerfall im südlichen Afrika."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("by IVE"))))
	// Java JLanguageTool runs disambiguation IGNORE_WHYS_AND_WHEREFORES → ignore_spelling on "whys"
	// (not surface invent TokenException). Simulate that Java-visible mark here.
	require.Equal(t, 0, len(rule.Match(withIgnoreSpelling("Never mind the whys and wherefores.", "whys"))))

	// checkSimpleReplaceRule twin: 1 match, 1 suggestion, exact string
	checkSimple := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "matches for %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()), "repl count for %q", sentence)
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0], "replacement for %q", sentence)
		// Java getMessage / getShort (match carries base AbstractSimpleReplaceRule pointer)
		require.Equal(t, "Possible spelling mistake found.", matches[0].GetMessage(), "message for %q", sentence)
		require.Equal(t, "Spelling mistake", matches[0].GetShortMessage(), "short for %q", sentence)
		base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
		require.True(t, ok, "rule type for %q", sentence)
		require.Equal(t, ContractionSpellingRuleID, base.GetID())
	}

	// at the beginning of a sentence
	checkSimple("Wasnt this great", "Wasn't")
	checkSimple("YOURE WRONG", "YOU'RE")
	checkSimple("Dont do this", "Don't")
	// Java has commented-out: checkSimpleReplaceRule("Ive been doing", "I've");
	// inside sentence
	checkSimple("It wasnt me", "wasn't")
	checkSimple("You neednt do this", "needn't")
	checkSimple("I know Im wrong", "I'm")

	// two suggestions (Whereve → Where've|Wherever)
	matches := rule.Match(languagetool.AnalyzePlain("Whereve you are"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 2, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, "Where've", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Wherever", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, "Possible spelling mistake found.", matches[0].GetMessage())
}

func TestContractionSpellingRule_FailClosedWithoutDisambig(t *testing.T) {
	rule := NewContractionSpellingRule(nil)
	// Without ignore_spelling from disambiguation, "whys" is in contractions.txt (Java data).
	matches := rule.Match(languagetool.AnalyzePlain("Never mind the whys and wherefores."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "why's", matches[0].GetSuggestedReplacements()[0])
}

// Java ContractionSpellingRule ctor metadata: ID, TYPOS, Misspelling, URL, example pair, case-sensitive, no lemmas.
func TestContractionSpellingRule_Metadata(t *testing.T) {
	rule := NewContractionSpellingRule(nil)
	require.Equal(t, ContractionSpellingRuleID, rule.GetID())
	require.Equal(t, "EN_CONTRACTION_SPELLING", rule.GetID())
	require.Equal(t, "Spelling of English contractions", rule.GetDescription())
	require.Equal(t, "https://languagetool.org/insights/post/grammar-contractions/", rule.GetURL())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	require.True(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	require.Equal(t, "Spelling mistake", rule.ShortMsg)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "We <marker>havent</marker> earned anything.", inc[0].GetExample())
	require.Equal(t, []string{"haven't"}, inc[0].GetCorrections())
	require.Equal(t, "We <marker>haven't</marker> earned anything.", rule.GetCorrectExamples()[0].GetExample())
}

// withIgnoreSpelling simulates Java disambiguation IGNORE_SPELLING on a surface token.
func withIgnoreSpelling(text, surface string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == surface {
			tok.IgnoreSpelling()
		}
	}
	return sent
}
