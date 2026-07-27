package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicHomophonesCheckRuleTest.java
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

// Java @Before: rule = new ArabicHomophonesRule(TestTools.getEnglishMessages());
//               lt = new JLanguageTool(new Arabic());
func arabicHomophonesFixture(t *testing.T) (*ArabicHomophonesRule, *languagetool.JLanguageTool) {
	t.Helper()
	return NewArabicHomophonesRule(nil), languagetool.NewJLanguageTool("ar")
}

// Java ArabicHomophonesCheckRuleTest.testRule full matrix.
func TestArabicHomophonesCheckRule_Rule(t *testing.T) {
	rule, lt := arabicHomophonesFixture(t)

	// incorrect sentences:
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("ضن"))))
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("حاضر"))))
	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("حض"))))

	// FIXME (Java, inactive): AbstractSimpleReplaceRule2 doesn't support lemma checking
	// assertEquals(1, rule.match(langTool.getAnalyzedSentence("الحاضر")).length);
	// Do not invent that assertion.
}

// Field twin of ArabicHomophonesRule constructor + overrides (beyond Java unit test).
func TestArabicHomophonesRule_Metadata(t *testing.T) {
	rule := NewArabicHomophonesRule(nil)

	require.Equal(t, ArabicHomophonesRuleID, rule.GetID())
	require.Equal(t, "AR_HOMOPHONES_REPLACE", rule.GetID())
	require.Equal(t, "كلمات متشابهة لفظا للتوضيح، يرجى التحقق منها مثل تشابه الظاء والضاد.", rule.GetDescription())
	require.Equal(t, "كلمات متشابهة لفظا يرجى التحقق منها", rule.GetShort())
	require.Equal(t, "قل $suggestions", rule.GetMessage())
	// Java getSuggestionsSeparator: " أو  " (space + أو + two spaces)
	require.Equal(t, " أو  ", rule.GetSuggestionsSeparator())
	require.Equal(t, " أو  ", rule.SuggestionsSeparator)
	require.Equal(t, "ar", rule.GetLocaleCode())
	require.Equal(t, ArabicHomophonesFileNames, rule.GetFileNames())
	require.Equal(t, []string{"ar/homophones.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java: Categories.CONFUSED_WORDS; Rule default ITS Uncategorized
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "CONFUSED_WORDS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())
	// Java does not setUrl / setDefaultOff
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// example pair: ضن → ظن
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>ضن</marker>", inc[0].GetExample())
	require.Equal(t, []string{"ظن"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>ظن</marker>", corr[0].GetExample())

	// Java-visible createMatch message/short for ضن hit (homophones.txt: ضن=ظن)
	lt := languagetool.NewJLanguageTool("ar")
	matches := rule.Match(lt.GetAnalyzedSentence("ضن"))
	require.Len(t, matches, 1)
	require.Equal(t, "كلمات متشابهة لفظا يرجى التحقق منها", matches[0].GetShortMessage())
	require.Equal(t,
		"قل <suggestion>ظن</suggestion>",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, ArabicHomophonesRuleID, base.GetID())
	// CONFUSED_WORDS category visible on match rule
	require.NotNil(t, base.GetCategory())
	require.Equal(t, "CONFUSED_WORDS", base.GetCategory().GetID().String())
}
