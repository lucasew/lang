package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicSimpleReplaceRuleTest.java
//
// Analysis: Java uses new JLanguageTool(Languages.getLanguageForShortCode("ar")).getAnalyzedSentence.
// Go uses arAnalyze / NewJLanguageTool("ar").GetAnalyzedSentence (same surface/token path).
// AbstractSimpleReplaceRule2 only inspects token surfaces, whitespaceBefore, and
// positions — not POS tags or lemmas.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java @Before: rule = new ArabicSimpleReplaceRule(TestTools.getMessages("ar"));
//               lt = new JLanguageTool(Languages.getLanguageForShortCode("ar"));
func arabicSimpleReplaceFixture(t *testing.T) (*ArabicSimpleReplaceRule, *languagetool.JLanguageTool) {
	t.Helper()
	return NewArabicSimpleReplaceRule(nil), languagetool.NewJLanguageTool("ar")
}

// Java ArabicSimpleReplaceRuleTest.testRule full matrix.
func TestArabicSimpleReplaceRule_Rule(t *testing.T) {
	rule, lt := arabicSimpleReplaceFixture(t)

	// correct:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("عبد الله"))))

	// incorrect sentences:
	// عبدالله → عبد الله
	matches := rule.Match(lt.GetAnalyzedSentence("عبدالله"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "عبد الله", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("يافطة"))))
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("المائة"))))
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("الذى"))))
}

// Field twin of ArabicSimpleReplaceRule constructor + overrides (beyond Java unit test).
func TestArabicSimpleReplaceRule_Metadata(t *testing.T) {
	rule := NewArabicSimpleReplaceRule(nil)

	require.Equal(t, ArabicSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "AR_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "قاعدة تطابق الكلمات التي يجب تجنبها وتقترح تصويبا لها", rule.GetDescription())
	// Java getShort: double space between أن and يقال
	require.Equal(t, "خطأ، يفضل أن  يقال:", rule.GetShort())
	require.Equal(t, "قل $suggestions", rule.GetMessage())
	// Java getSuggestionsSeparator: " أو  " (space + أو + two spaces)
	require.Equal(t, " أو  ", rule.GetSuggestionsSeparator())
	require.Equal(t, " أو  ", rule.SuggestionsSeparator)
	require.Equal(t, "ar", rule.GetLocaleCode())
	require.Equal(t, ArabicSimpleReplaceFileNames, rule.GetFileNames())
	require.Equal(t, []string{"ar/replaces.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	// Java isDictionaryBasedSpellingRule() → false
	require.False(t, rule.IsDictionaryBasedSpellingRule())

	// Java: Categories.CONFUSED_WORDS; Rule default ITS Uncategorized
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "CONFUSED_WORDS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())
	// Java does not setUrl / setDefaultOff
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// example pair: الى → إلى
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>الى</marker>", inc[0].GetExample())
	require.Equal(t, []string{"إلى"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>إلى</marker>", corr[0].GetExample())

	// Java-visible createMatch message/short for عبدالله hit
	lt := languagetool.NewJLanguageTool("ar")
	matches := rule.Match(lt.GetAnalyzedSentence("عبدالله"))
	require.Len(t, matches, 1)
	require.Equal(t, "خطأ، يفضل أن  يقال:", matches[0].GetShortMessage())
	require.Equal(t,
		"قل <suggestion>عبد الله</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, ArabicSimpleReplaceRuleID, base.GetID())
	// CONFUSED_WORDS category visible on match rule
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "CONFUSED_WORDS", base.GetCategory().GetID().String())
}
