package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.LongSentenceRuleTest — Java-visible outcomes only.
//
// Java setup:
//
//	JLanguageTool lt = new JLanguageTool(TestTools.getDemoLanguage());
//	LongSentenceRule rule = new LongSentenceRule(TestTools.getEnglishMessages(), new UserConfig(), 40);
//	rule.match(lt.analyzeText(input))  // assert length, fromPos, toPos
//
// Go: demoLT() (xx + SimpleSentenceTokenizer) + MatchList(lt.Analyze).

func engMessagesLongSentence() map[string]string {
	// MessagesBundle_en keys used by LongSentenceRule.
	return map[string]string{
		"long_sentence_rule_msg2": "This sentence is over {0} words long. Consider splitting it up, as shorter sentences make the text easier to read.",
		"long_sentence_rule_desc": "Readability: sentence over {0} words",
		"guiLongSentencesText":    "Maximum sentences length (5 - 100 words):",
	}
}

// matchLongSentence ports rule.match(lt.analyzeText(text)).
func matchLongSentence(rule *LongSentenceRule, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

func TestLongSentenceRule_Match(t *testing.T) {
	// Java: new LongSentenceRule(TestTools.getEnglishMessages(), new UserConfig(), 40)
	rule := NewLongSentenceRuleWithUserConfig(engMessagesLongSentence(), languagetool.NewUserConfig(), 40)
	require.Equal(t, LongSentenceRuleID, rule.GetID())
	require.True(t, rule.HasTag(TagPicky))
	require.Equal(t, ITSStyle, rule.GetLocQualityIssueType())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, 0, rule.MinToCheckParagraph())
	require.Equal(t, 40, rule.MaxWords)

	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())

	assertNoMatch := func(input string) {
		t.Helper()
		m := matchLongSentence(rule, lt, input)
		require.Equal(t, 0, len(m), "assertNoMatch %q", input)
	}
	assertMatch := func(input string, from, to int) {
		t.Helper()
		m := matchLongSentence(rule, lt, input)
		require.Equal(t, 1, len(m), "assertMatch %q got %d", input, len(m))
		require.Equal(t, from, m[0].GetFromPos(), "fromPos for %q", input)
		require.Equal(t, to, m[0].GetToPos(), "toPos for %q", input)
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

	// test quotes (QUOTED_SENT_END / indexOfQuote exemptions)
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

	// no quotes
	assertMatch("The quote When days grow dark and nights grow dreary, we can be thankful that our God combines in "+
		"his nature a creative synthesis of love and justice which will lead us through life’s dark valleys and into "+
		"sunlit pathways of hope and fulfillment (p. 9) refers to God’s nature as a combination of love and justice.",
		0, 249)

	// wrong or mismatched quotes: opening quote, with no matching closing quote
	// (QUOTED_SENT_END still finds ?!. + quote sequence; or words stay inside open quote)
	assertNoMatch("The quote “When days grow dark and nights grow dreary, we can be thankful that our God combines in " +
		"his nature a creative synthesis of love and justice which will lead us through life’s dark valleys and into " +
		"sunlit pathways of hope and fulfillment» (p. 9) refers to God’s nature as a combination of love and justice.")

	// Java: shortRule = new LongSentenceRule(..., new UserConfig(), 6)
	shortRule := NewLongSentenceRuleWithUserConfig(engMessagesLongSentence(), languagetool.NewUserConfig(), 6)
	assertNoMatchShort := func(input string) {
		t.Helper()
		m := matchLongSentence(shortRule, lt, input)
		require.Equal(t, 0, len(m), "short no match %q", input)
	}
	assertMatchShort := func(input string, from, to int) {
		t.Helper()
		m := matchLongSentence(shortRule, lt, input)
		require.Equal(t, 1, len(m), "short match %q got %d", input, len(m))
		require.Equal(t, from, m[0].GetFromPos(), "short fromPos for %q", input)
		require.Equal(t, to, m[0].GetToPos(), "short toPos for %q", input)
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

func TestLongSentenceRule_UserConfigMaxWords(t *testing.T) {
	// Java: userConfig.getConfigValueByID(getId()) overrides maxWords.
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID(LongSentenceRuleID, []any{10})
	rule := NewLongSentenceRuleWithUserConfig(engMessagesLongSentence(), uc, 40)
	require.Equal(t, 10, rule.MaxWords)
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 10, opts[0].DefaultValue)
	require.Equal(t, 5, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
	require.Equal(t, "Maximum sentences length (5 - 100 words):", opts[0].ConfigureText)
}

func TestLongSentenceRule_MessageKeys(t *testing.T) {
	rule := NewLongSentenceRule(engMessagesLongSentence(), 40)
	require.Equal(t, "Readability: sentence over 40 words", rule.GetDescription())
	require.Contains(t, rule.GetMessage(), "40")
}
