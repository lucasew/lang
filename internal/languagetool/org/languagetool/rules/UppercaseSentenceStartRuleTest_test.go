package rules

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.UppercaseSentenceStartRuleTest

func TestUppercaseSentenceStartRule_Rule(t *testing.T) {
	rule := NewUppercaseSentenceStartRule(map[string]string{
		"incorrect_case":          "This sentence does not start with an uppercase letter",
		"desc_uppercase_sentence": "Checks that a sentence starts with an uppercase letter",
		"category_case":           "Capitalization",
	}, "en")

	analyze := func(s string) []*languagetool.AnalyzedSentence {
		if s == "" {
			return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("")}
		}
		if strings.Contains(s, ". ") || strings.Contains(s, "! ") || strings.Contains(s, "? ") ||
			strings.Contains(s, ".\n") || strings.Contains(s, "?\n") || strings.Contains(s, "!\n") {
			return languagetool.SplitAndAnalyze(s)
		}
		return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}
	}

	assertGood := func(s string) {
		t.Helper()
		m := rule.MatchList(analyze(s))
		require.Equal(t, 0, len(m), "assertGood %q got %d", s, len(m))
	}

	assertGood("this")
	assertGood("a) This is a test sentence.")
	assertGood("iv. This is a test sentence...")
	assertGood("\"iv. This is a test sentence...\"")
	assertGood("\u00bbiv. This is a test sentence...")
	assertGood("This")
	assertGood("This is")
	assertGood("This is a test sentence")
	assertGood("")
	assertGood("http://www.languagetool.org")
	assertGood("eBay can be at sentence start in lowercase.")
	assertGood("\u00bfEsto es una pregunta?")
	assertGood("\u00bfEsto es una pregunta?, \u00bfy esto?")
	assertGood("\u00f8 This is a test sentence with a wrong bullet character.")

	matches := rule.MatchList(analyze("this is a test sentence."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 4, matches[0].GetToPos())

	matches2 := rule.MatchList(analyze("this!"))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 0, matches2[0].GetFromPos())
	require.Equal(t, 4, matches2[0].GetToPos())

	require.Equal(t, 1, len(rule.MatchList(analyze("'this is a sentence'."))))
	require.Equal(t, 1, len(rule.MatchList(analyze("\"this is a sentence.\""))))
	require.Equal(t, 1, len(rule.MatchList(analyze("\u201ethis is a sentence."))))
	require.Equal(t, 1, len(rule.MatchList(analyze("\u00abthis is a sentence."))))
	require.Equal(t, 1, len(rule.MatchList(analyze("\u2018this is a sentence."))))
	require.Equal(t, 1, len(rule.MatchList(analyze("\u00bfesto es una pregunta?"))))

	// second sentence starts with lowercase "y"
	sents := languagetool.SplitAndAnalyze("\u00bfEsto es una pregunta? \u00bfy esto?")
	require.Equal(t, 1, len(rule.MatchList(sents)))
}
