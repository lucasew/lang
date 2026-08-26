package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/AmericanReplaceRuleTest.java
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(en-US).getAnalyzedSentence for this rule’s surface matching.
// AmericanReplaceRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official replace.txt
// multiword keys (e.g. "fish fingers") tokenize the same under WordTokenizer as EN.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Twin of AmericanReplaceRuleTest.testRule
func TestAmericanReplaceRule_Rule(t *testing.T) {
	// Java: new AmericanReplaceRule(TestTools.getMessages("en"), "/en/en-US/replace.txt")
	//       lt = new JLanguageTool(Languages.getLanguageForShortCode("en-US"))
	rule := NewAmericanReplaceRule(nil, AmericanReplaceFilePath)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Buy some gasoline."))))

	// incorrect sentences: expression
	checkSimpleReplaceRule(t, rule, "I love fish fingers.", "fish sticks")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	matches := rule.Match(languagetool.AnalyzePlain("I love fish fingers."))
	require.Len(t, matches, 1)
	require.Equal(t, "British word", matches[0].GetShortMessage())
	require.Equal(t,
		"'fish fingers' is a common British expression. Consider using expressions more common to American English.",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language en-US
	wantID := tools.ToId(AmericanSimpleReplaceRuleID+"_"+"fish fingers", "en-US")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "EN_US_SIMPLE_REPLACE_FISH_FINGERS", sr.GetID())
	require.Equal(t, "British words easily confused in American English: fish fingers", sr.GetDescription())
	require.Equal(t, 7, matches[0].GetFromPos())
	require.Equal(t, 19, matches[0].GetToPos())
}

// checkSimpleReplaceRule ports AmericanReplaceRuleTest.checkSimpleReplaceRule.
func checkSimpleReplaceRule(t *testing.T, rule *AmericanReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count wile checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of AmericanReplaceRule constructor + overrides (not in Java test class,
// but required for sector field-by-field parity).
func TestAmericanReplaceRule_Metadata(t *testing.T) {
	rule := NewAmericanReplaceRule(nil, AmericanReplaceFilePath)

	require.Equal(t, AmericanSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "EN_US_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "British words easily confused in American English: $match", rule.GetDescription())
	require.Equal(t, "British word", rule.GetShort())
	require.Equal(t, "'$match' is a common British expression. Consider using expressions more common to American English.", rule.GetMessage())
	require.Equal(t, "en-US", rule.GetLocaleCode())
	require.Equal(t, []string{AmericanReplaceFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/en/en-US/replace.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Are baked <marker>crisps</marker> healthy?", inc[0].GetExample())
	require.Equal(t, []string{"chips"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "Are baked <marker>chips</marker> healthy?", corr[0].GetExample())
}
