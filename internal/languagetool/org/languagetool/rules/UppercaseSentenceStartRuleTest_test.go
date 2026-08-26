package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.UppercaseSentenceStartRuleTest — Java-visible outcomes only.
//
// Java setup:
//   rule = new UppercaseSentenceStartRule(TestTools.getEnglishMessages(),
//           TestTools.getDemoLanguage(), Example.wrong(...), Example.fixed(...));
//   lt   = new JLanguageTool(TestTools.getDemoLanguage());
//   assertGood → rule.match(lt.analyzeText(s)).length == 0
//   incorrect → rule.match(lt.analyzeText(...)) with match count + from/to when asserted.
// Go: demoLT() (xx + SimpleSentenceTokenizer + ignoredCharacters) + MatchList(lt.Analyze).

func engMessagesUppercase() map[string]string {
	// MessagesBundle_en keys used by UppercaseSentenceStartRule.
	return map[string]string{
		"incorrect_case":          "This sentence does not start with an uppercase letter",
		"desc_uppercase_sentence": "Checks that a sentence starts with an uppercase letter",
		"category_case":           "Capitalization",
	}
}

func uppercaseSentenceStartRule() *UppercaseSentenceStartRule {
	// Java: language = Demo (short code "xx"); messages = English.
	return NewUppercaseSentenceStartRule(engMessagesUppercase(), "xx")
}

// matchUSSR ports rule.match(lt.analyzeText(text)) for this TextLevelRule.
func matchUSSR(rule *UppercaseSentenceStartRule, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

// assertGoodUSSR ports UppercaseSentenceStartRuleTest.assertGood.
func assertGoodUSSR(t *testing.T, rule *UppercaseSentenceStartRule, lt *languagetool.JLanguageTool, s string) {
	t.Helper()
	matches := matchUSSR(rule, lt, s)
	require.Equal(t, 0, len(matches), "assertGood %q got %d matches", s, len(matches))
}

func TestUppercaseSentenceStartRule_Rule(t *testing.T) {
	rule := uppercaseSentenceStartRule()
	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())

	// correct sentences:
	assertGoodUSSR(t, rule, lt, "this")
	assertGoodUSSR(t, rule, lt, "a) This is a test sentence.")
	assertGoodUSSR(t, rule, lt, "iv. This is a test sentence...")
	assertGoodUSSR(t, rule, lt, "\"iv. This is a test sentence...\"")
	assertGoodUSSR(t, rule, lt, "\u00bbiv. This is a test sentence...")
	assertGoodUSSR(t, rule, lt, "This")
	assertGoodUSSR(t, rule, lt, "This is")
	assertGoodUSSR(t, rule, lt, "This is a test sentence")
	assertGoodUSSR(t, rule, lt, "")
	assertGoodUSSR(t, rule, lt, "http://www.languagetool.org")
	assertGoodUSSR(t, rule, lt, "eBay can be at sentence start in lowercase.")
	assertGoodUSSR(t, rule, lt, "\u00bfEsto es una pregunta?")
	assertGoodUSSR(t, rule, lt, "\u00bfEsto es una pregunta?, \u00bfy esto?")
	assertGoodUSSR(t, rule, lt, "\u00f8 This is a test sentence with a wrong bullet character.")

	// incorrect sentences (Java: match counts + from/to where asserted):
	matches := matchUSSR(rule, lt, "this is a test sentence.")
	require.Equal(t, 1, len(matches))
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 4, matches[0].GetToPos())

	matches2 := matchUSSR(rule, lt, "this!")
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 0, matches2[0].GetFromPos())
	require.Equal(t, 4, matches2[0].GetToPos())

	require.Equal(t, 1, len(matchUSSR(rule, lt, "'this is a sentence'.")))
	require.Equal(t, 1, len(matchUSSR(rule, lt, "\"this is a sentence.\"")))
	require.Equal(t, 1, len(matchUSSR(rule, lt, "\u201ethis is a sentence.")))
	require.Equal(t, 1, len(matchUSSR(rule, lt, "\u00abthis is a sentence.")))
	require.Equal(t, 1, len(matchUSSR(rule, lt, "\u2018this is a sentence.")))
	require.Equal(t, 1, len(matchUSSR(rule, lt, "\u00bfesto es una pregunta?")))
	require.Equal(t, 1, len(matchUSSR(rule, lt, "\u00bfEsto es una pregunta? \u00bfy esto?")))
}
