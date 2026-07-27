package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicDarjaRuleTest.java
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

// Java @Before:
//
//	rule = new ArabicDarjaRule(TestTools.getMessages("ar"));
//	lt = new JLanguageTool(Languages.getLanguageForShortCode("ar"));
func arabicDarjaFixture(t *testing.T) (*ArabicDarjaRule, *languagetool.JLanguageTool) {
	t.Helper()
	return NewArabicDarjaRule(nil), languagetool.NewJLanguageTool("ar")
}

// Java ArabicDarjaRuleTest.testRule.
func TestArabicDarjaRule_Rule(t *testing.T) {
	rule, lt := arabicDarjaFixture(t)

	// correct sentences:
	// assertEquals(0, rule.match(lt.getAnalyzedSentence("إن شاء")).length);
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("إن شاء"))))

	// incorrect sentences:
	// RuleMatch[] matches = rule.match(lt.getAnalyzedSentence("طرشي"));
	// assertEquals(1, matches.length);
	// assertEquals("فلفل حلو", matches[0].getSuggestedReplacements().get(0));
	matches := rule.Match(lt.GetAnalyzedSentence("طرشي"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "فلفل حلو", matches[0].GetSuggestedReplacements()[0])

	// assertEquals(1, rule.match(lt.getAnalyzedSentence("فايدة")).length);
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("فايدة"))))
}

// Field twin of ArabicDarjaRule constructor + overrides (beyond Java unit test).
func TestArabicDarjaRule_Metadata(t *testing.T) {
	rule := NewArabicDarjaRule(nil)

	require.Equal(t, ArabicDarjaRuleID, rule.GetID())
	require.Equal(t, "AR_DARJA_REPLACE", rule.GetID())
	require.Equal(t, "كلمات بديلة للكلمات العامية أو الأجنبية", rule.GetDescription())
	require.Equal(t, "كلمات بديلة للكلمات العامية أو الأجنبية", rule.GetShort())
	// Java getMessage: double space after عامية
	require.Equal(t,
		"الكلمة عامية  أو أجنبية يفضل أن يقال $suggestions",
		rule.GetMessage())
	// Java getSuggestionsSeparator: " أو  " (space + أو + TWO trailing spaces)
	require.Equal(t, " أو  ", rule.GetSuggestionsSeparator())
	require.Equal(t, " أو  ", rule.SuggestionsSeparator)
	require.Equal(t, "ar", rule.GetLocaleCode())
	require.Equal(t, ArabicDarjaFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/ar/darja.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java: Categories.STYLE; setLocQualityIssueType(ITSIssueType.LocaleViolation)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())
	// Java does not setUrl / setDefaultOff
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// example pair: طرشي → فلفل حلو
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>طرشي</marker>", inc[0].GetExample())
	require.Equal(t, []string{"فلفل حلو"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>فلفل حلو</marker>", corr[0].GetExample())

	// Java-visible createMatch message/short for طرشي hit
	// (darja.txt: طرشي=فلفل حلو)
	lt := languagetool.NewJLanguageTool("ar")
	matches := rule.Match(lt.GetAnalyzedSentence("طرشي"))
	require.Len(t, matches, 1)
	require.Equal(t, "فلفل حلو", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "كلمات بديلة للكلمات العامية أو الأجنبية", matches[0].GetShortMessage())
	require.Equal(t,
		"الكلمة عامية  أو أجنبية يفضل أن يقال <suggestion>فلفل حلو</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, ArabicDarjaRuleID, base.GetID())
	// STYLE category + LocaleViolation issue type visible on match rule
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "STYLE", base.GetCategory().GetID().String())
	require.Equal(t, rules.ITSLocaleViolation, base.GetLocQualityIssueType())
}
