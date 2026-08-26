package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/BritishReplaceRuleTest.java
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(en-GB).getAnalyzedSentence for this rule’s surface matching.
// BritishReplaceRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official replace.txt
// multiword keys (e.g. "cotton candy") tokenize the same under WordTokenizer as EN.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Twin of BritishReplaceRuleTest.testRule
func TestBritishReplaceRule_Rule(t *testing.T) {
	// Java: new BritishReplaceRule(TestTools.getMessages("en"), "/en/en-GB/replace.txt")
	//       lt = new JLanguageTool(Languages.getLanguageForShortCode("en-GB"))
	rule := NewBritishReplaceRule(nil, BritishReplaceFilePath)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Buy some petrol."))))

	// incorrect sentences:

	// at the beginning of a sentence (Romanian replace rule is case-sensitive)
	checkBritishSimpleReplaceRule(t, rule, "Diapers for sale!", "Nappies")
	// inside sentence
	checkBritishSimpleReplaceRule(t, rule, "We have some diapers for sale.", "nappies")
	// expression
	checkBritishSimpleReplaceRule(t, rule, "Everyone loves cotton candy.", "candy floss")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	matches := rule.Match(languagetool.AnalyzePlain("Everyone loves cotton candy."))
	require.Len(t, matches, 1)
	require.Equal(t, "American word", matches[0].GetShortMessage())
	require.Equal(t,
		"'cotton candy' is a common American expression. Consider using expressions more common to British English.",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language en-GB
	wantID := tools.ToId(BritishSimpleReplaceRuleID+"_"+"cotton candy", "en-GB")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "EN_GB_SIMPLE_REPLACE_COTTON_CANDY", sr.GetID())
	require.Equal(t, "American words easily confused in British English: cotton candy", sr.GetDescription())
	require.Equal(t, 15, matches[0].GetFromPos())
	require.Equal(t, 27, matches[0].GetToPos())
}

// checkBritishSimpleReplaceRule ports BritishReplaceRuleTest.checkSimpleReplaceRule.
func checkBritishSimpleReplaceRule(t *testing.T, rule *BritishReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count wile checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of BritishReplaceRule constructor + overrides (not in Java test class,
// but required for sector field-by-field parity).
func TestBritishReplaceRule_Metadata(t *testing.T) {
	rule := NewBritishReplaceRule(nil, BritishReplaceFilePath)

	require.Equal(t, BritishSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "EN_GB_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "American words easily confused in British English: $match", rule.GetDescription())
	require.Equal(t, "American word", rule.GetShort())
	require.Equal(t, "'$match' is a common American expression. Consider using expressions more common to British English.", rule.GetMessage())
	require.Equal(t, "en-GB", rule.GetLocaleCode())
	require.Equal(t, []string{BritishReplaceFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/en/en-GB/replace.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "We can produce <marker>drapes</marker> of any size or shape from a choice of over 500 different fabrics.", inc[0].GetExample())
	require.Equal(t, []string{"curtains"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "We can produce <marker>curtains</marker> of any size or shape from a choice of over 500 different fabrics.", corr[0].GetExample())
}
