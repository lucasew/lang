package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicDiacriticsCheckRuleTest.java
//
// Analysis: Java uses new JLanguageTool(new Arabic()).getAnalyzedSentence.
// Go uses NewJLanguageTool("ar").GetAnalyzedSentence (same surface/token path).
// AbstractSimpleReplaceRule2 only inspects token surfaces, whitespaceBefore, and
// positions — not POS tags or lemmas.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java @Before:
//
//	rule = new ArabicDiacriticsRule(TestTools.getEnglishMessages());
//	lt = new JLanguageTool(new Arabic());
func arabicDiacriticsFixture(t *testing.T) (*ArabicDiacriticsRule, *languagetool.JLanguageTool) {
	t.Helper()
	return NewArabicDiacriticsRule(nil), languagetool.NewJLanguageTool("ar")
}

// Java ArabicDiacriticsCheckRuleTest.testRule.
// Java only asserts match count (does NOT assert suggestion text).
func TestArabicDiacriticsCheckRule_Rule(t *testing.T) {
	rule, lt := arabicDiacriticsFixture(t)

	// incorrect sentences:
	// assertEquals(1, rule.match(lt.getAnalyzedSentence("تجربة")).length);
	matches := rule.Match(lt.GetAnalyzedSentence("تجربة"))
	require.Equal(t, 1, len(matches))
}

// Field twin of ArabicDiacriticsRule constructor + overrides (beyond Java unit test).
func TestArabicDiacriticsRule_Metadata(t *testing.T) {
	rule := NewArabicDiacriticsRule(nil)

	require.Equal(t, ArabicDiacriticsRuleID, rule.GetID())
	require.Equal(t, "AR_DIACRITICS_REPLACE", rule.GetID())
	require.Equal(t, "كلمات مشكولة للتوضيح", rule.GetDescription())
	require.Equal(t, "كلمات يستحسن أن تشكّل لتصحيح نطقها", rule.GetShort())
	require.Equal(t,
		"'$match' كلمة يشيع نطقها نطقا خاطئا لذا نقترح تشكيلها كالآتي: $suggestions",
		rule.GetMessage())
	// Java getSuggestionsSeparator: " أو " (space + أو + one trailing space)
	require.Equal(t, " أو ", rule.GetSuggestionsSeparator())
	require.Equal(t, " أو ", rule.SuggestionsSeparator)
	require.Equal(t, "ar", rule.GetLocaleCode())
	require.Equal(t, ArabicDiacriticsFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/ar/diacritics.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java: Categories.STYLE; setLocQualityIssueType(ITSIssueType.Style)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	// Java does not setUrl / setDefaultOff
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// example pair: تجربة → تجرِِبة (Java double-kasra quirk in example, not data)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>تجربة</marker>", inc[0].GetExample())
	require.Equal(t, []string{"تجرِِبة"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>تجرِِبة</marker>", corr[0].GetExample())

	// Java-visible createMatch message/short for تجربة hit
	// (diacritics.txt: تجربة=تجرِبة single kasra)
	lt := languagetool.NewJLanguageTool("ar")
	matches := rule.Match(lt.GetAnalyzedSentence("تجربة"))
	require.Len(t, matches, 1)
	require.Equal(t, "تجرِبة", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "كلمات يستحسن أن تشكّل لتصحيح نطقها", matches[0].GetShortMessage())
	require.Equal(t,
		"'تجربة' كلمة يشيع نطقها نطقا خاطئا لذا نقترح تشكيلها كالآتي: <suggestion>تجرِبة</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, ArabicDiacriticsRuleID, base.GetID())
	// STYLE category + Style issue type visible on match rule
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "STYLE", base.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, base.GetLocQualityIssueType())
}
