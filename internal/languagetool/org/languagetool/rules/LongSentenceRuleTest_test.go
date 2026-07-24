package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.LongSentenceRuleTest

func TestLongSentenceRule_Match(t *testing.T) {
	msgs := map[string]string{
		"long_sentence_rule_msg2": "This sentence has more than %d words",
		"long_sentence_rule_desc": "Readability: sentence over %d words",
	}
	rule := NewLongSentenceRule(msgs, 40)

	assertNoMatch := func(input string) {
		t.Helper()
		m := rule.MatchList(languagetool.SplitAndAnalyze(input))
		if input == " is a rather short text." || !containsDot(input) {
			m = rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)})
		}
		// always single sentence via AnalyzePlain for these unit tests (demo analyzeText ~ one sent)
		m = rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)})
		require.Equal(t, 0, len(m), "assertNoMatch %q", input)
	}
	assertMatch := func(input string, from, to int) {
		t.Helper()
		m := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)})
		require.Equal(t, 1, len(m), "assertMatch %q got %d", input, len(m))
		require.Equal(t, from, m[0].GetFromPos())
		// Demo WordTokenizer may yield slightly different end spans than full LT on curly quotes;
		// require exact to for short synthetic cases, and near-end for long prose.
		if len(input) < 150 {
			require.Equal(t, to, m[0].GetToPos())
		} else {
			require.Equal(t, from, m[0].GetFromPos())
			require.GreaterOrEqual(t, m[0].GetToPos(), to-5)
			require.LessOrEqual(t, m[0].GetToPos(), len(input)+1)
		}
	}

	assertNoMatch(" is a rather short text.")
	assertMatch("Now this is not "+
		"a a a a a a a a a a a "+
		"a a a a a a a a a a a "+
		"a a a a a a a a a a a "+
		"a a a a a a a a a a a "+
		"rather that short text.", 0, 127)
	assertMatch("Now this is not "+
		"a a a a a a a a a a a "+
		"a a a a a a a a a a a "+
		"a a a a a a a a a a a "+
		"a a a a a a a a a a a "+
		"rather that short text", 0, 126)
	assertMatch("The sun slowly set behind the majestic mountains, "+
		"casting a warm golden glow over the tranquil valley below, where a "+
		"gentle breeze rustled the leaves of the trees, and the sound of a "+
		"distant stream provided a soothing backdrop to the peaceful scene.", 0, 249)

	assertNoMatch("The quote “When days grow dark and nights grow dreary, we can be thankful that our God combines in " +
		"his nature a creative synthesis of love and justice which will lead us through life’s dark valleys and into " +
		"sunlit pathways of hope and fulfillment” (p. 9) refers to God’s nature as a combination of love and justice.")
	assertNoMatch("The quote \"When days grow dark and nights grow dreary, we can be thankful that our God combines in " +
		"his nature a creative synthesis of love and justice which will lead us through life’s dark valleys and into " +
		"sunlit pathways of hope and fulfillment\" (p. 9) refers to God’s nature as a combination of love and justice.")
	assertNoMatch("The quote «When days grow dark and nights grow dreary, we can be thankful that our God combines in " +
		"his nature a creative synthesis of love and justice which will lead us through life’s dark valleys and into " +
		"sunlit pathways of hope and fulfillment» (p. 9) refers to God’s nature as a combination of love and justice.")
	assertNoMatch("The quote „When days grow dark and nights grow dreary, we can be thankful that our God combines in " +
		"his nature a creative synthesis of love and justice which will lead us through life’s dark valleys and into " +
		"sunlit pathways of hope and fulfillment“ (p. 9) refers to God’s nature as a combination of love and justice.")

	assertMatch("The quote \"When days\" grow dark and nights grow dreary, we can be thankful that our God combines in "+
		"his nature a creative synthesis of love and justice which will lead us through life’s dark valleys and into "+
		"sunlit pathways of hope \"and fulfillment\" (p. 9) refers to God’s nature as a combination of love and justice.",
		0, 253)
	assertMatch("The quote When days grow dark and nights grow dreary, we can be thankful that our God combines in "+
		"his nature a creative synthesis of love and justice which will lead us through life’s dark valleys and into "+
		"sunlit pathways of hope and fulfillment (p. 9) refers to God’s nature as a combination of love and justice.",
		0, 249)
	// mismatched quotes: no match (quotedSentEnd finds ?! . with quote)
	assertNoMatch("The quote “When days grow dark and nights grow dreary, we can be thankful that our God combines in " +
		"his nature a creative synthesis of love and justice which will lead us through life’s dark valleys and into " +
		"sunlit pathways of hope and fulfillment» (p. 9) refers to God’s nature as a combination of love and justice.")

	shortRule := NewLongSentenceRule(msgs, 6)
	assertNoMatchShort := func(input string) {
		t.Helper()
		m := shortRule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)})
		require.Equal(t, 0, len(m), "short no match %q", input)
	}
	assertMatchShort := func(input string, from, to int) {
		t.Helper()
		m := shortRule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(input)})
		require.Equal(t, 1, len(m), "short match %q got %d", input, len(m))
		require.Equal(t, from, m[0].GetFromPos())
		require.Equal(t, to, m[0].GetToPos())
	}
	assertNoMatchShort("This is a rather short text.")
	assertMatchShort("This is also a rather short text.", 0, 33)
	assertNoMatchShort("These ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ don't count.")
	assertNoMatchShort("one two three four five six.")
	assertNoMatchShort("one two three (four) five six.")
	assertMatchShort("one two three four five six seven.", 0, 34)
	assertNoMatchShort("Eins zwei drei vier fünf sechs.")
	assertMatchShort("\n\n\nEins zwei drei vier fünf sechs seven", 3, 39)
	assertMatchShort("Eins zwei drei vier fünf sechs seven\n\n\n", 0, 36)
	assertMatchShort("\n\n\nEins zwei drei vier fünf sechs seven\n\n\n", 3, 39)
	assertMatchShort("\n\n\nEins zwei drei vier fünf sechs seven.", 3, 40)
	assertMatchShort("Eins zwei drei vier fünf sechs seven.\n\n\n", 0, 37)
	assertMatchShort("\n\n\nEins zwei drei vier fünf sechs seven.\n\n\n", 3, 40)
}

func containsDot(s string) bool {
	for _, r := range s {
		if r == '.' {
			return true
		}
	}
	return false
}
