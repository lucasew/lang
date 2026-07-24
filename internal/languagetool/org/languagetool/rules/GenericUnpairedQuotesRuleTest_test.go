package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.GenericUnpairedQuotesRuleTest.

func quotesRule() *GenericUnpairedQuotesRule {
	return NewGenericUnpairedQuotesRule(map[string]string{
		"unpaired_brackets":      "Unpaired bracket, expected %s",
		"desc_unpaired_brackets": "Unpaired brackets",
	},
		[]string{"\u00bb", "\u00ab", "\"", "'", "\u203a", "\u2039"},
		[]string{"\u00ab", "\u00bb", "\"", "'", "\u2039", "\u203a"},
	)
}

// analyzeForQuotes uses whole-text AnalyzePlain (local positions). WhitespaceBefore
// is set by AnalyzePlain; multi-sentence paragraph structure is encoded as \n/\n\n
// tokens. Matches Java FakeLanguage+JLanguageTool for these unit cases.
func analyzeForQuotes(input string) []*languagetool.AnalyzedSentence {
	return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)}
}

func assertQuoteMatches(t *testing.T, expected int, input string) {
	t.Helper()
	rule := quotesRule()
	got := len(rule.MatchList(analyzeForQuotes(input)))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestGenericUnpairedQuotesRule_Rule(t *testing.T) {
	assertQuoteMatches(t, 0, "This is \u00bbcorrect\u00ab.")
	assertQuoteMatches(t, 0, "This is \u00abcorrect\u00bb.")
	assertQuoteMatches(t, 0, "\u00bbCorrect\u00ab\n\u00bbAnd \u203ahere\u2039 it ends.\u00ab")
	assertQuoteMatches(t, 0, "\u00abCorrect\u00bb\n\u00abAnd \u2039here\u203a it ends.\u00bb")
	assertQuoteMatches(t, 0, "\u00bbCorrect. This is more than one sentence.\u00ab")
	assertQuoteMatches(t, 0, "\u00bbCorrect. This is more than one sentence.\u00ab\n\u00bbAnd \u203ahere\u2039 it ends.\u00ab")
	assertQuoteMatches(t, 0, "\u00bbCorrect\u00ab\n\n\u00bbAnd here it ends.\u00ab\n\nMore text.")
	assertQuoteMatches(t, 0, "\u00bbCorrect, he said. This is the next sentence.\u00ab Here's another sentence.")
	assertQuoteMatches(t, 0, "\u00bbCorrect\u00ab, he said.\n\n\u00bbThis is the next sentence.\u00ab \u00bbHere's another sentence.\u00ab")
	assertQuoteMatches(t, 0, "\u00bbCorrect\u00ab, he said. \u00bbThis is the next sentence.\u00ab\n\u00bbHere's another sentence.\u00ab")
	assertQuoteMatches(t, 0, "This \u00bbis also \u203acorrect\u2039\u00ab.")
	assertQuoteMatches(t, 0, "Good.\n\nThis \u00bbis also \u203acorrect\u2039\u00ab.")
	assertQuoteMatches(t, 0, "Good.\n\n\nThis \u00bbis also \u203acorrect\u2039\u00ab.")
	assertQuoteMatches(t, 0, "Good.\n\n\n\nThis \u00bbis also \u203acorrect\u2039\u00ab.")

	assertQuoteMatches(t, 0, "This is \"correct\".")
	assertQuoteMatches(t, 0, "This is 'correct'.")
	assertQuoteMatches(t, 0, "\"Correct\"\n\"And 'here' it ends.\"")
	assertQuoteMatches(t, 0, "'Correct'\n'And \"here\" it ends.'")
	assertQuoteMatches(t, 0, "This \"is also 'correct'\".")
	assertQuoteMatches(t, 0, "This isn't \"incorrect\".")
	assertQuoteMatches(t, 0, "\"This isn't incorrect.\"")
	assertQuoteMatches(t, 0, "The screen is 20\" wide.")
	assertQuoteMatches(t, 0, "\"The screen is 20\" wide.\"")

	assertQuoteMatches(t, 1, "This is not correct\u00ab")
	assertQuoteMatches(t, 1, "This is \u00bbnot correct.")
	assertQuoteMatches(t, 1, "This is \u00bbnot correct")
	assertQuoteMatches(t, 1, "This is \u00bbnot an error yet\n\nBut now it has become one")
	assertQuoteMatches(t, 1, "This is correct.\n\n\u00bbBut this is not.")
	assertQuoteMatches(t, 1, "This is correct.\n\nBut this is not\u00ab")
	assertQuoteMatches(t, 1, "\u00bbThis is correct\u00ab\n\nBut this is not\u00ab")
	assertQuoteMatches(t, 1, "\u00bbThis is correct\u00ab\n\nBut this \u00bbis\u00ab not\u00ab")
	assertQuoteMatches(t, 1, "This is not correct. No matter if it's more than one sentence\u00ab")
	assertQuoteMatches(t, 1, "\u00bbThis is not correct. No matter if it's more than one sentence")
	assertQuoteMatches(t, 1, "Correct, he said. This is the next sentence.\u00ab Here's another sentence.")
	assertQuoteMatches(t, 1, "\u00bbCorrect, he said. This is the next sentence. Here's another sentence.")
	assertQuoteMatches(t, 1, "\u00bbCorrect, he said. This is the next sentence.\n\nHere's another sentence.")
	assertQuoteMatches(t, 1, "\u00bbCorrect, he said. This is the next sentence.\n\n\n\nHere's another sentence.")

	assertQuoteMatches(t, 2, "\u00bbCorrect\u00ab\n\u00bbAnd \u00bbhere\u00ab it ends with two matches.\u00ab")
	assertQuoteMatches(t, 2, "\u00bbCorrect. This is more than one sentence.\u00ab\n\u00bbAnd \u00bbhere\u00ab it ends.\u00ab")
	assertQuoteMatches(t, 2, "Good.\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertQuoteMatches(t, 2, "Good.\n\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
	assertQuoteMatches(t, 2, "Good.\n\n\n\nThis \u00bbis also \u00bbcorrect\u00ab\u00ab.")
}

func TestGenericUnpairedQuotesRule_RuleMatchPositions(t *testing.T) {
	rule := quotesRule()
	match1 := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("This \u00bbis a test.")})
	require.NotEmpty(t, match1)
	require.Equal(t, 5, match1[0].GetFromPos())
	require.Equal(t, 6, match1[0].GetToPos())

	text2 := "This.\nSome stuff.\nIt \u00bbis a test."
	match2 := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(text2)})
	require.NotEmpty(t, match2)
	require.Equal(t, 21, match2[0].GetFromPos())
	require.Equal(t, 22, match2[0].GetToPos())

	// NBSP counts as one UTF-16 unit: Th + nbsp + is + space + »
	match3 := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("Th\u00ADis \u00bbis a test.")})
	require.NotEmpty(t, match3)
	require.Equal(t, 6, match3[0].GetFromPos())
	require.Equal(t, 7, match3[0].GetToPos())
}
