package km

// Outcome twin for org.languagetool.rules.km.KhmerSimpleReplaceRule.
// No dedicated Java *Test for this type — cases derived from official
// /km/coherency.txt + AbstractSimpleReplaceRule2 behavior (same pattern as
// ACCEPTed BE SimpleReplaceRule / EL ReplaceHomonymsRule twins).
//
// Analysis: AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. AnalyzePlain
// (WordTokenizer + SENT_START) is sufficient for single-token keys in
// coherency.txt (no space-separated multiword keys in that resource).

import (
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func khmerSimpleReplaceFixture(t *testing.T) *KhmerSimpleReplaceRule {
	t.Helper()
	return NewKhmerSimpleReplaceRule(nil)
}

// Example from coherency.txt: សំដី=សម្ដី
func TestKhmerSimpleReplaceRule_Rule(t *testing.T) {
	rule := khmerSimpleReplaceFixture(t)

	// incorrect: សំដី → សម្ដី
	matches := rule.Match(languagetool.AnalyzePlain("សំដី"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, "សម្ដី", matches[0].GetSuggestedReplacements()[0])
	// Java createMatch: short = getShort(); message rebuild when per-entry msg null
	// getMessage() has no $match/$suggestions → literal with leading/trailing spaces
	require.Equal(t, "Consider following the spelling of Chuon Nath", matches[0].GetShortMessage())
	require.Equal(t, " Consider following the spelling of Chuon Nath ", matches[0].GetMessage())
	// span covers the whole wrong surface (UTF-16 code units, same as Java positions)
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, len(utf16.Encode([]rune("សំដី"))), matches[0].GetToPos())
	// Java does not call useSubRuleSpecificIds()
	_, isSub := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSub, "no sub-rule specific id expected")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok)
	require.Equal(t, KhmerSimpleReplaceRuleID, base.GetID())

	// correct spelling (right-hand side of សំដី=សម្ដី) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("សម្ដី"))))
}

// Multi-key line: មិត្តសំឡាញ់|មិត្តសំលាញ់|មិត្រសម្លាញ់=មិត្តសម្លាញ់
func TestKhmerSimpleReplaceRule_MultiKeyLine(t *testing.T) {
	rule := khmerSimpleReplaceFixture(t)
	want := "មិត្តសម្លាញ់"
	for _, wrong := range []string{"មិត្តសំឡាញ់", "មិត្តសំលាញ់", "មិត្រសម្លាញ់"} {
		matches := rule.Match(languagetool.AnalyzePlain(wrong))
		require.Equal(t, 1, len(matches), "wrong=%q", wrong)
		require.Equal(t, want, matches[0].GetSuggestedReplacements()[0], "wrong=%q", wrong)
		require.Equal(t, 0, matches[0].GetFromPos())
		require.Equal(t, len(utf16.Encode([]rune(wrong))), matches[0].GetToPos(), "wrong=%q", wrong)
	}
	// correct form → 0
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(want))))
}

// Longer single-token key (no spaces in coherency.txt multiword keys):
// ពាក្យសំដី=ពាក្យសម្ដី
func TestKhmerSimpleReplaceRule_CompoundKey(t *testing.T) {
	rule := khmerSimpleReplaceFixture(t)
	matches := rule.Match(languagetool.AnalyzePlain("ពាក្យសំដី"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "ពាក្យសម្ដី", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, len(utf16.Encode([]rune("ពាក្យសំដី"))), matches[0].GetToPos())
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("ពាក្យសម្ដី"))))
}

// Another multi-key line with two wrongs: កែលម្អ|កែលំអរ=កែលំអ
func TestKhmerSimpleReplaceRule_TwoKeyLine(t *testing.T) {
	rule := khmerSimpleReplaceFixture(t)
	want := "កែលំអ"
	for _, wrong := range []string{"កែលម្អ", "កែលំអរ"} {
		matches := rule.Match(languagetool.AnalyzePlain(wrong))
		require.Equal(t, 1, len(matches), "wrong=%q", wrong)
		require.Equal(t, want, matches[0].GetSuggestedReplacements()[0], "wrong=%q", wrong)
	}
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(want))))
}

// Field twin of KhmerSimpleReplaceRule constructor + overrides.
func TestKhmerSimpleReplaceRule_Metadata(t *testing.T) {
	rule := NewKhmerSimpleReplaceRule(nil)

	require.Equal(t, KhmerSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "KM_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Words or groups of words that are incorrect or obsolete", rule.GetDescription())
	require.Equal(t, "Consider following the spelling of Chuon Nath", rule.GetShort())
	// Java getMessage spaces preserved exactly
	require.Equal(t, " Consider following the spelling of Chuon Nath ", rule.GetMessage())
	// Java getSuggestionsSeparator override
	require.Equal(t, " or ", rule.GetSuggestionsSeparator())
	require.Equal(t, " or ", rule.SuggestionsSeparator)
	// Java: private static final Locale KM_LOCALE = new Locale("km");
	require.Equal(t, "km", rule.GetLocaleCode())
	require.Equal(t, KhmerSimpleReplaceFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/km/coherency.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java does not override getCaseSensitivy → CaseSensitivy.CI
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	// Java AbstractSimpleReplaceRule2 + Khmer ctor: Categories.MISC; Rule default ITS Uncategorized
	// (unlike BE SimpleReplaceRule which sets Misspelling — Khmer does not)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())
	// Java does not setUrl / setDefaultOff / addExamplePair
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())
	require.Nil(t, rule.GetIncorrectExamples())
	require.Nil(t, rule.GetCorrectExamples())
}
