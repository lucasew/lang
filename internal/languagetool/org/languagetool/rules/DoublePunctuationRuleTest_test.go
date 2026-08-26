package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.DoublePunctuationRuleTest — Java-visible outcomes only.
//
// Java setup:
//   rule = new DoublePunctuationRule(TestTools.getEnglishMessages());
//   lt   = new JLanguageTool(TestTools.getDemoLanguage());

func doublePunctuationRule() *DoublePunctuationRule {
	// Messages mirror MessagesBundle_en.properties keys used by the rule.
	return NewDoublePunctuationRule(map[string]string{
		"two_dots":            "Two consecutive dots",
		"two_commas":          "Two consecutive commas",
		"double_dots_short":   "Two consecutive dots",
		"double_commas_short": "Two consecutive commas",
		"desc_double_punct":   "Use of two consecutive dots or commas",
	})
}

// assertDPMatches ports rule.match(lt.getAnalyzedSentence(text)).length
func assertDPMatches(t *testing.T, rule *DoublePunctuationRule, lt *languagetool.JLanguageTool, text string, expected int) {
	t.Helper()
	n := len(rule.Match(lt.GetAnalyzedSentence(text)))
	require.Equal(t, expected, n, "text=%q", text)
}

func TestDoublePunctuationRule_Rule(t *testing.T) {
	rule := doublePunctuationRule()
	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())

	// correct sentences:
	assertDPMatches(t, rule, lt, "This is a test sentence...", 0)
	assertDPMatches(t, rule, lt, "Это тестовое предложение?..", 0)
	assertDPMatches(t, rule, lt, "Это тестовое предложение!.. ", 0)
	assertDPMatches(t, rule, lt, "This is a test sentence... More stuff....", 0)
	assertDPMatches(t, rule, lt, "This is a test sentence..... More stuff....", 0)
	assertDPMatches(t, rule, lt, "This, is, a test sentence.", 0)
	assertDPMatches(t, rule, lt, "The path is ../data/vtest.avi", 0)
	assertDPMatches(t, rule, lt, "The path is ..\\data\\vtest.avi", 0)
	assertDPMatches(t, rule, lt, "Something … … ..", 0)
	assertDPMatches(t, rule, lt, "Something … … ... …", 0)
	assertDPMatches(t, rule, lt, "Something … … .... …", 0)
	assertDPMatches(t, rule, lt, "Something … … .. …", 0)
	assertDPMatches(t, rule, lt, "Something ……..", 0)

	// errors:
	matches1 := rule.Match(lt.GetAnalyzedSentence("This,, is a test sentence."))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 4, matches1[0].GetFromPos())
	require.Equal(t, 6, matches1[0].GetToPos())

	matches2 := rule.Match(lt.GetAnalyzedSentence("This is a test sentence.. Another sentence"))
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 23, matches2[0].GetFromPos())
	require.Equal(t, 25, matches2[0].GetToPos())
}
