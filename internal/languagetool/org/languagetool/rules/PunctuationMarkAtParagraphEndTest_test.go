package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.PunctuationMarkAtParagraphEndTest — Java-visible outcomes only.
//
// Java setup (setUpRule):
//
//	JLanguageTool lt = new JLanguageTool(TestTools.getDemoLanguage());
//	for (Rule rule : lt.getAllRules()) { lt.disableRule(rule.getId()); }
//	PunctuationMarkAtParagraphEnd rule = new PunctuationMarkAtParagraphEnd(
//	    TestTools.getEnglishMessages(), TestTools.getDemoLanguage());
//	rule.setTags(Collections.emptyList());
//	lt.addRule(rule);
//	assertEquals(n, lt.check(input).size());
//
// Go: demoLT() (xx + SimpleSentenceTokenizer) + MatchList(lt.Analyze) for the same text-level path
// as rule.match(lt.analyzeText(...)) / lt.check with only this rule enabled.

func engMessagesParaEnd() map[string]string {
	// MessagesBundle_en keys used by PunctuationMarkAtParagraphEnd.
	return map[string]string{
		"punctuation_mark_paragraph_end_msg":  "Please add a punctuation mark at the end of paragraph.",
		"punctuation_mark_paragraph_end_desc": "No punctuation mark at the end of paragraph",
		"category_punctuation":                "Punctuation",
	}
}

func setUpParaEndRule() (*PunctuationMarkAtParagraphEnd, *languagetool.JLanguageTool) {
	// Java: new PunctuationMarkAtParagraphEnd(TestTools.getEnglishMessages(), TestTools.getDemoLanguage())
	rule := NewPunctuationMarkAtParagraphEnd(engMessagesParaEnd())
	// Java: rule.setTags(Collections.emptyList());
	rule.SetTags(nil)
	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())
	return rule, lt
}

// matchParaEnd ports rule.match(lt.analyzeText(text)) / lt.check with only this rule.
func matchParaEnd(rule *PunctuationMarkAtParagraphEnd, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

func assertParaEndMatches(t *testing.T, rule *PunctuationMarkAtParagraphEnd, lt *languagetool.JLanguageTool, expected int, input string) {
	t.Helper()
	got := len(matchParaEnd(rule, lt, input))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestPunctuationMarkAtParagraphEnd_Rule(t *testing.T) {
	rule, lt := setUpParaEndRule()
	require.Equal(t, "PUNCTUATION_PARAGRAPH_END", rule.GetID())
	require.False(t, rule.IsDefaultOff())
	require.Equal(t, 0, rule.MinToCheckParagraph())
	// setTags(empty) clears picky
	require.False(t, rule.HasTag(TagPicky))
	// Fresh ctor keeps picky
	require.True(t, NewPunctuationMarkAtParagraphEnd(engMessagesParaEnd()).HasTag(TagPicky))

	assertParaEndMatches(t, rule, lt, 0, "A paragraph.\n2. Some headline\n\n(a) A new sentence.")
	assertParaEndMatches(t, rule, lt, 0, "A paragraph.\n\n2. Some headline\n\n(a) A new sentence.")
	assertParaEndMatches(t, rule, lt, 0, "A paragraph.\n2.2.1 Some headline\n\n(a) A new sentence.")
	assertParaEndMatches(t, rule, lt, 0, "A paragraph.\n\n2.2.1 Some headline\n\n(a) A new sentence.")
	assertParaEndMatches(t, rule, lt, 0, "2. This is an item in a list")
	assertParaEndMatches(t, rule, lt, 0, "2.2.2. This is an item in a list")
	assertParaEndMatches(t, rule, lt, 0, "a) This is an item in a list")
	assertParaEndMatches(t, rule, lt, 0, "a.) This is an item in a list")
	assertParaEndMatches(t, rule, lt, 0, "\u2713 This is an item in a list")
	assertParaEndMatches(t, rule, lt, 0, "* This is an item in a list")
	assertParaEndMatches(t, rule, lt, 0, "This is a test sentence.")
	assertParaEndMatches(t, rule, lt, 0, "This is a test headline")
	assertParaEndMatches(t, rule, lt, 0, "This is a test sentence. This is a link: http://example.com") // no error because of colon
	assertParaEndMatches(t, rule, lt, 1, "This is a test sentence. It can be found at http://example.com/foobar")
	assertParaEndMatches(t, rule, lt, 1, "This is a test sentence. And this is a second test sentence")
	assertParaEndMatches(t, rule, lt, 1, "\"This is a test sentence. And this is a second test sentence")
	assertParaEndMatches(t, rule, lt, 0, "This is a test sentence. And this is a second test sentence.")
	assertParaEndMatches(t, rule, lt, 0, "B. v. – Beschluss vom")
	assertParaEndMatches(t, rule, lt, 1,
		"This is a test sentence.\nAnd this is a second test sentence. Here is a dot missing")
	assertParaEndMatches(t, rule, lt, 0,
		"This is a test sentence.\nAnd this is a second test sentence. Here is a dot missing.")
	assertParaEndMatches(t, rule, lt, 0,
		"This is a sentence. Another one: https://languagetool.org/foo\n\nAnother sentence\n")
}

func TestPunctuationMarkAtParagraphEnd_DefaultActive(t *testing.T) {
	// Java 3-arg: defaultActive false → setDefaultOff()
	off := NewPunctuationMarkAtParagraphEndDefaultActive(engMessagesParaEnd(), false)
	require.True(t, off.IsDefaultOff())
	on := NewPunctuationMarkAtParagraphEndDefaultActive(engMessagesParaEnd(), true)
	require.False(t, on.IsDefaultOff())
}
