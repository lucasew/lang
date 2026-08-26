package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.PunctuationMarkAtParagraphEnd2Test — Java-visible outcomes only.
//
// Java setup:
//
//	JLanguageTool lt = new JLanguageTool(TestTools.getDemoLanguage());
//	for (Rule rule : lt.getAllRules()) { lt.disableRule(rule.getId()); }
//	PunctuationMarkAtParagraphEnd2 rule = new PunctuationMarkAtParagraphEnd2(
//	    TestTools.getEnglishMessages(), TestTools.getDemoLanguage());
//	rule.setDefaultOn();
//	lt.addRule(rule);
//	assertEquals(n, lt.check(input).size());
//
// Go: demoLT() (xx + SimpleSentenceTokenizer) + MatchList(lt.Analyze) for the same
// text-level path as rule.match(lt.analyzeText(...)) / lt.check with only this rule.

func engMessagesParaEnd2() map[string]string {
	// MessagesBundle_en keys used by PunctuationMarkAtParagraphEnd2.
	return map[string]string{
		"punctuation_mark_paragraph_end_msg":  "Please add a punctuation mark at the end of paragraph.",
		"punctuation_mark_paragraph_end_desc": "No punctuation mark at the end of paragraph",
		"category_punctuation":                "Punctuation",
	}
}

func setUpParaEnd2Rule() (*PunctuationMarkAtParagraphEnd2, *languagetool.JLanguageTool) {
	// Java: new PunctuationMarkAtParagraphEnd2(TestTools.getEnglishMessages(), TestTools.getDemoLanguage())
	rule := NewPunctuationMarkAtParagraphEnd2(engMessagesParaEnd2())
	// Java: rule.setDefaultOn();
	rule.SetDefaultOn()
	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())
	return rule, lt
}

// matchParaEnd2 ports rule.match(lt.analyzeText(text)) / lt.check with only this rule.
func matchParaEnd2(rule *PunctuationMarkAtParagraphEnd2, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

func assertParaEnd2Matches(t *testing.T, rule *PunctuationMarkAtParagraphEnd2, lt *languagetool.JLanguageTool, expected int, input string) {
	t.Helper()
	got := len(matchParaEnd2(rule, lt, input))
	require.Equal(t, expected, got, "input=%q", input)
}

func TestPunctuationMarkAtParagraphEnd2_Test(t *testing.T) {
	rule, lt := setUpParaEnd2Rule()
	require.Equal(t, "PUNCTUATION_PARAGRAPH_END2", rule.GetID())
	require.False(t, rule.IsDefaultOff()) // after setDefaultOn
	require.Equal(t, 0, rule.MinToCheckParagraph())
	require.Equal(t, -1, rule.EstimateContextForSureMatch())
	// Fresh ctor is default-off (Java setDefaultOff in constructor)
	require.True(t, NewPunctuationMarkAtParagraphEnd2(engMessagesParaEnd2()).IsDefaultOff())
	require.Equal(t, "No punctuation mark at the end of paragraph", rule.GetDescription())

	// Java PunctuationMarkAtParagraphEnd2Test.test()
	assertParaEnd2Matches(t, rule, lt, 0, "2. This is an item in a list")
	assertParaEnd2Matches(t, rule, lt, 0, "2.2.2. This is an item in a list")
	assertParaEnd2Matches(t, rule, lt, 0, "This is a test.")
	assertParaEnd2Matches(t, rule, lt, 0, "This is a test")             // too short
	assertParaEnd2Matches(t, rule, lt, 0, "This is a really nice test") // no error, might not be finished
	assertParaEnd2Matches(t, rule, lt, 1, "This is a really nice test, and it has enough tokens\n")
	assertParaEnd2Matches(t, rule, lt, 1, "This is a really nice test, and it has enough tokens\n\n")
	assertParaEnd2Matches(t, rule, lt, 0, "\"This is a really nice test, and it has enough tokens.\"\n\n")
	assertParaEnd2Matches(t, rule, lt, 0, "\"This is a really nice test, and it has enough tokens\"\n")   // not really clear whether the rule should match here...
	assertParaEnd2Matches(t, rule, lt, 0, "\"This is a really nice test, and it has enough tokens\"\n\n") // not really clear whether the rule should match here...
	assertParaEnd2Matches(t, rule, lt, 0, "This is a test.\n\nRegards,\nJim")
	assertParaEnd2Matches(t, rule, lt, 0, "This is a test.\n\nRegards,\n\nJim")
	assertParaEnd2Matches(t, rule, lt, 0, "This is a test.\n\nKind Regards,\nJim")
	assertParaEnd2Matches(t, rule, lt, 0, "This is a test.\n\nKind Regards,\n\nJim")
	assertParaEnd2Matches(t, rule, lt, 0, "This is a test.\n\nKind Regards,\n\nJim Tester")
	assertParaEnd2Matches(t, rule, lt, 0, "This is a test.\n\nKind Regards,\n\nJim van Tester")
	assertParaEnd2Matches(t, rule, lt, 0, "This is headline-style text")
	assertParaEnd2Matches(t, rule, lt, 0, "This is headline-style text.")
	assertParaEnd2Matches(t, rule, lt, 0, "This is headline-style text. If it gets longer, a dot is needed.")
	assertParaEnd2Matches(t, rule, lt, 1, "This is headline-style text. If it gets longer, a dot is needed")
	assertParaEnd2Matches(t, rule, lt, 0, "This is a test\n\nKind Regards,\n\nJim van Tester") // too short to find missing dot
	assertParaEnd2Matches(t, rule, lt, 1, "This is a really nice test, and it has enough tokens\n\nKind Regards,\n\nJim van Tester")
}

func TestPunctuationMarkAtParagraphEnd2_DefaultOff(t *testing.T) {
	// Java constructor always calls setDefaultOff().
	off := NewPunctuationMarkAtParagraphEnd2(engMessagesParaEnd2())
	require.True(t, off.IsDefaultOff())
	off.SetDefaultOn()
	require.False(t, off.IsDefaultOff())
	off.SetDefaultOff()
	require.True(t, off.IsDefaultOff())
}
