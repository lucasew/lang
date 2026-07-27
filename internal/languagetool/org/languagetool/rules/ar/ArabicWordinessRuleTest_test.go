package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicWordinessRuleTest.java
//
// Analysis: Java uses new JLanguageTool(Languages.getLanguageForShortCode("ar")).getAnalyzedSentence.
// Go uses NewJLanguageTool("ar").GetAnalyzedSentence (same surface/token path).
// AbstractSimpleReplaceRule2 only inspects token surfaces, whitespaceBefore, and
// positions — not POS tags or lemmas.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java @Before: TestTools.disableAllRulesExcept(lt, ArabicWordinessRule.AR_WORDINESS_REPLACE);
//
//	lt = new JLanguageTool(Languages.getLanguageForShortCode("ar"));
//
// assertError creates a fresh ArabicWordinessRule(TestTools.getEnglishMessages())
// and matches against lt.getAnalyzedSentence(s).
func arabicWordinessFixture(t *testing.T) (*ArabicWordinessRule, *languagetool.JLanguageTool) {
	t.Helper()
	return NewArabicWordinessRule(nil), languagetool.NewJLanguageTool("ar")
}

// Java ArabicWordinessRuleTest.testRule.
func TestArabicWordinessRule_Rule(t *testing.T) {
	rule, lt := arabicWordinessFixture(t)

	// errors:
	// assertError("وأخيرا وليس آخرا");
	matches := rule.Match(lt.GetAnalyzedSentence("وأخيرا وليس آخرا"))
	require.Equal(t, 1, len(matches))
}

// Field twin of ArabicWordinessRule constructor + overrides (beyond Java unit test).
func TestArabicWordinessRule_Metadata(t *testing.T) {
	rule := NewArabicWordinessRule(nil)

	require.Equal(t, ArabicWordinessRuleID, rule.GetID())
	require.Equal(t, "AR_WORDINESS_REPLACE", rule.GetID())
	require.Equal(t, "2. حشو(تعبير فيه تكرار)", rule.GetDescription())
	require.Equal(t, "حشو (تعبير فيه تكرار)", rule.GetShort())
	require.Equal(t,
		"'$match' تعبير فيه حشو يفضل أن يقال $suggestions",
		rule.GetMessage())
	// Java getSuggestionsSeparator: " أو " (space + أو + one trailing space)
	require.Equal(t, " أو ", rule.GetSuggestionsSeparator())
	require.Equal(t, " أو ", rule.SuggestionsSeparator)
	require.Equal(t, "ar", rule.GetLocaleCode())
	require.Equal(t, ArabicWordinessFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/ar/wordiness.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java: Categories.REDUNDANCY; setLocQualityIssueType(ITSIssueType.Style)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "REDUNDANCY", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	// Java does not setUrl / setDefaultOff
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// example pair: هناك خطأ في العبارة → في العبارة خطأ
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>هناك خطأ في العبارة</marker>", inc[0].GetExample())
	require.Equal(t, []string{"في العبارة خطأ"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>في العبارة خطأ</marker>", corr[0].GetExample())

	// Java-visible createMatch message/short for وأخيرا وليس آخرا hit
	// (wordiness.txt: وأخيرا وليس آخرا=وأخيرا)
	lt := languagetool.NewJLanguageTool("ar")
	matches := rule.Match(lt.GetAnalyzedSentence("وأخيرا وليس آخرا"))
	require.Len(t, matches, 1)
	require.Equal(t, "وأخيرا", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "حشو (تعبير فيه تكرار)", matches[0].GetShortMessage())
	require.Equal(t,
		"'وأخيرا وليس آخرا' تعبير فيه حشو يفضل أن يقال <suggestion>وأخيرا</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, ArabicWordinessRuleID, base.GetID())
	// REDUNDANCY category + Style issue type visible on match rule
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "REDUNDANCY", base.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, base.GetLocQualityIssueType())
}
