package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.SentenceWhitespaceRuleTest — Java-visible outcomes only.
//
// Java setup:
//
//	rule = new SentenceWhitespaceRule(TestTools.getEnglishMessages());
//	lt   = new JLanguageTool(new FakeLanguage());
//	lt.addRule(rule);
//	assertGood / assertBad → languageTool.check(text).size() is 0 or 1
//
// FakeLanguage extends Language with short code "yy", empty relevant rules, and
// createDefaultSentenceTokenizer() → SimpleSentenceTokenizer (segment-simple.srx).
// Go: "yy" + SimpleSentenceTokenizer + MatchList(lt.Analyze) (TextLevelRule path
// equivalent to check with only this rule enabled).

func engMessagesSentenceWhitespace() map[string]string {
	// MessagesBundle_en keys used by SentenceWhitespaceRule.
	return map[string]string{
		"addSpaceBetweenSentences":        "Add a space between sentences",
		"missing_space_between_sentences": "Missing space between sentences",
	}
}

// fakeLanguageLT ports new JLanguageTool(new FakeLanguage()).
// FakeLanguage → Language.createDefaultSentenceTokenizer → SimpleSentenceTokenizer.
func fakeLanguageLT() *languagetool.JLanguageTool {
	lt := languagetool.NewJLanguageTool("yy")
	lt.SentenceTokenizer = tokenizers.NewSimpleSentenceTokenizer()
	return lt
}

// matchSentenceWhitespace ports lt.check(text) with only SentenceWhitespaceRule enabled:
// analyzeText (SimpleSentenceTokenizer) then rule.match(sentences).
func matchSentenceWhitespace(rule *SentenceWhitespaceRule, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

// assertGoodSWR ports SentenceWhitespaceRuleTest.assertGood → check size 0.
func assertGoodSWR(t *testing.T, rule *SentenceWhitespaceRule, lt *languagetool.JLanguageTool, text string) {
	t.Helper()
	matches := matchSentenceWhitespace(rule, lt, text)
	require.Equal(t, 0, len(matches), "assertGood %q got %d matches", text, len(matches))
}

// assertBadSWR ports SentenceWhitespaceRuleTest.assertBad → check size 1.
func assertBadSWR(t *testing.T, rule *SentenceWhitespaceRule, lt *languagetool.JLanguageTool, text string) {
	t.Helper()
	matches := matchSentenceWhitespace(rule, lt, text)
	require.Equal(t, 1, len(matches), "assertBad %q got %d matches", text, len(matches))
}

func TestSentenceWhitespaceRule_Match(t *testing.T) {
	rule := NewSentenceWhitespaceRule(engMessagesSentenceWhitespace())
	lt := fakeLanguageLT()

	require.Equal(t, "SENTENCE_WHITESPACE", rule.GetID())
	require.Equal(t, 0, rule.MinToCheckParagraph())

	// Java assertGood:
	assertGoodSWR(t, rule, lt, "This is a text. And there's the next sentence.")
	assertGoodSWR(t, rule, lt, "This is a text! And there's the next sentence.")
	assertGoodSWR(t, rule, lt, "This is a text\nAnd there's the next sentence.")
	assertGoodSWR(t, rule, lt, "This is a text\n\nAnd there's the next sentence.")

	// Java assertBad:
	assertBadSWR(t, rule, lt, "This is a text.And there's the next sentence.")
	assertBadSWR(t, rule, lt, "This is a text!And there's the next sentence.")
	assertBadSWR(t, rule, lt, "This is a text?And there's the next sentence.")
}
