package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicRedundancyRuleTest.java
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

// Java @Before: rule = new ArabicRedundancyRule(TestTools.getMessages("ar"));
//
//	lt = new JLanguageTool(Languages.getLanguageForShortCode("ar"));
func arabicRedundancyFixture(t *testing.T) (*ArabicRedundancyRule, *languagetool.JLanguageTool) {
	t.Helper()
	return NewArabicRedundancyRule(nil), languagetool.NewJLanguageTool("ar")
}

// Java ArabicRedundancyRuleTest.testRule full matrix.
func TestArabicRedundancyRule_Rule(t *testing.T) {
	rule, lt := arabicRedundancyFixture(t)

	// incorrect sentences:
	matches := rule.Match(lt.GetAnalyzedSentence("سوف لن"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "لن", matches[0].GetSuggestedReplacements()[0])
}

// Field twin of ArabicRedundancyRule constructor + overrides (beyond Java unit test).
func TestArabicRedundancyRule_Metadata(t *testing.T) {
	rule := NewArabicRedundancyRule(nil)

	require.Equal(t, ArabicRedundancyRuleID, rule.GetID())
	require.Equal(t, "AR_REDUNDANCY_REPLACE", rule.GetID())
	require.Equal(t, "1. تكرار (عام)", rule.GetDescription())
	require.Equal(t, "تكرار", rule.GetShort())
	require.Equal(t,
		"'$match' تعبير فيه تكرار.في بعض الحالات، يستحسن استعمال $suggestions",
		rule.GetMessage())
	// Java getSuggestionsSeparator: " أو " (space + أو + one trailing space)
	require.Equal(t, " أو ", rule.GetSuggestionsSeparator())
	require.Equal(t, " أو ", rule.SuggestionsSeparator)
	require.Equal(t, "ar", rule.GetLocaleCode())
	require.Equal(t, ArabicRedundancyFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/ar/redundancies.txt"}, rule.GetFileNames())
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

	// example pair: سوف لن → لن
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>سوف لن</marker>", inc[0].GetExample())
	require.Equal(t, []string{"لن"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>لن</marker>", corr[0].GetExample())

	// Java-visible createMatch message/short for سوف لن hit (redundancies.txt: سوف لن=لن)
	lt := languagetool.NewJLanguageTool("ar")
	matches := rule.Match(lt.GetAnalyzedSentence("سوف لن"))
	require.Len(t, matches, 1)
	require.Equal(t, "تكرار", matches[0].GetShortMessage())
	require.Equal(t,
		"'سوف لن' تعبير فيه تكرار.في بعض الحالات، يستحسن استعمال <suggestion>لن</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, ArabicRedundancyRuleID, base.GetID())
	// REDUNDANCY category + Style issue type visible on match rule
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "REDUNDANCY", base.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, base.GetLocQualityIssueType())
}
