package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.LongParagraphRuleTest — Java-visible outcomes only.
//
// Java setup:
//
//	Demo lang = new Demo();
//	LongParagraphRule rule = new LongParagraphRule(TestTools.getEnglishMessages(), lang, null, 6);
//	JLanguageTool lt = new JLanguageTool(lang);
//	rule.match(lt.analyzeText(...))
//
// Go: demoLT() (xx + SimpleSentenceTokenizer) + MatchList(lt.Analyze).

func engMessagesLongParagraph() map[string]string {
	// MessagesBundle_en keys used by LongParagraphRule.
	return map[string]string{
		"long_paragraph_rule_msg":  "This paragraph is over {0} words long here, consider revising it.",
		"long_paragraph_rule_desc": "Readability: paragraph over {0} words",
		"guiLongParagraphsText":    "Maximum paragraphs length (5 - {0} words):",
	}
}

// matchLongParagraph ports rule.match(lt.analyzeText(text)).
func matchLongParagraph(rule *LongParagraphRule, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

func TestLongParagraphRule_Rule(t *testing.T) {
	// Java: new LongParagraphRule(TestTools.getEnglishMessages(), lang, null, 6)
	rule := NewLongParagraphRule(engMessagesLongParagraph(), 6)
	require.Equal(t, "TOO_LONG_PARAGRAPH", rule.GetID())
	require.True(t, rule.IsDefaultOff())
	require.True(t, rule.IsOfficeDefaultOn())
	require.True(t, rule.HasTag(TagPicky))
	require.Equal(t, 0, rule.MinToCheckParagraph())
	require.Equal(t, 6, rule.MaxWords)

	lt := demoLT() // new JLanguageTool(new Demo())

	assertLen := func(text string, n int) {
		t.Helper()
		m := matchLongParagraph(rule, lt, text)
		require.Equal(t, n, len(m), "text=%q", text)
	}

	assertLen("This is a short paragraph.", 0)
	assertLen("This is only almost long paragraph by unit test standards.", 0)
	assertLen("Here's some text as a filler. This is a long paragraph by unit test standards.", 1)
	assertLen("Here's some text as a filler.  A test. A long paragraph by unit test standards.", 1)
	assertLen("Here's some text as a filler.  A test.\nNot a long paragraph.\nBecause of the line breaks.\n", 0)
	assertLen("- [ ] A test.\n- [ ] Not a long paragraph.\n- [ ] Because of the line breaks.\n- [ ] More text even.\n", 0)

	text1 := "This is a short paragraph.\n\nHere's some text as filler. This is a long paragraph by unit test standards."
	matches1 := matchLongParagraph(rule, lt, text1)
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 45, matches1[0].GetFromPos()) // "filler" span: maxWords-1..maxWords
	require.Equal(t, 54, matches1[0].GetToPos())

	text2 := "Here's some text as filler. This is a long paragraph by unit test standards.\n\nAnother paragraph.\n\nHere's some text as morefiller - this is a long paragraph by unit test standards."
	matches2 := matchLongParagraph(rule, lt, text2)
	require.Equal(t, 2, len(matches2))
	require.Equal(t, 17, matches2[0].GetFromPos()) // "filler"
	require.Equal(t, 26, matches2[0].GetToPos())
	require.Equal(t, 115, matches2[1].GetFromPos()) // "morefiller"
	require.Equal(t, 128, matches2[1].GetToPos())
}

func TestLongParagraphRule_UserConfigMaxWords(t *testing.T) {
	// Java: userConfig.getConfigValueByID(getId()) overrides maxWords.
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID(LongParagraphRuleID, []any{10})
	rule := NewLongParagraphRuleWithUserConfig(engMessagesLongParagraph(), uc, 6)
	require.Equal(t, 10, rule.MaxWords)
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 10, opts[0].DefaultValue)
}

func TestLongParagraphRule_DefaultMaxWords(t *testing.T) {
	// Java 3-arg: new LongParagraphRule(messages, lang, userConfig) → defaultWords=-1 → 220.
	rule := NewLongParagraphRule(engMessagesLongParagraph(), -1)
	require.Equal(t, 220, rule.MaxWords)
}
