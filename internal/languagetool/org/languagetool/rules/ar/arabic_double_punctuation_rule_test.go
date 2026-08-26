package ar

// Outcome twins for org.languagetool.rules.ar.ArabicDoublePunctuationRule.
// Java has no dedicated *Test.java for this type; behavior is DoublePunctuationRule
// with getCommaCharacter() = "،" (U+060C) and getId ARABIC_DOUBLE_PUNCTUATION.
// Upstream DoublePunctuationRuleTest covers base outcomes (dots / ASCII commas);
// here we twin AR-visible paths via real ar analysis (ArabicWordTokenizer splits on "،").

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func arabicDoublePunctMessages() map[string]string {
	// Mirror English MessagesBundle keys used by DoublePunctuationRule
	// (TestTools.getEnglishMessages in DoublePunctuationRuleTest).
	return map[string]string{
		"two_dots":            "Two consecutive dots",
		"two_commas":          "Two consecutive commas",
		"double_dots_short":   "Two consecutive dots",
		"double_commas_short": "Two consecutive commas",
		"desc_double_punct":   "Use of two consecutive dots or commas",
	}
}

// arAnalyze is defined in ArabicCommaWhitespaceRuleTest_test.go (same package).

func TestArabicDoublePunctuationRule_Metadata(t *testing.T) {
	rule := NewArabicDoublePunctuationRule(arabicDoublePunctMessages())
	require.Equal(t, "ARABIC_DOUBLE_PUNCTUATION", rule.GetID())
	require.Equal(t, "،", rule.GetCommaCharacter())
	require.Equal(t, "Use of two consecutive dots or commas", rule.GetDescription())
}

func TestArabicDoublePunctuationRule_Rule(t *testing.T) {
	rule := NewArabicDoublePunctuationRule(arabicDoublePunctMessages())
	assertMatches := func(text string, n int) {
		t.Helper()
		require.Equal(t, n, len(rule.Match(arAnalyze(text))), "text=%q", text)
	}

	// correct — single punctuation / ellipsis / non-Arabic comma
	assertMatches("نعم،", 0)
	assertMatches("نعم، لقد نجحنا", 0)
	assertMatches("هذه جملة تجريبية.", 0)
	// three dots (base rule: not exactly two consecutive dots)
	assertMatches("نعم...", 0)
	assertMatches("This is a test sentence...", 0)
	// ASCII comma is not getCommaCharacter() → no double-comma match
	assertMatches("نعم,,", 0)
	assertMatches("This,, is a test.", 0)
	// paths / Russian ?.. !.. (base DoublePunctuationRuleTest)
	assertMatches("The path is ../data/vtest.avi", 0)
	assertMatches("Это тестовое предложение?..", 0)
	assertMatches("Это тестовое предложение!.. ", 0)

	// errors — double Arabic comma (U+060C), glued (ArabicWordTokenizer splits each ،)
	assertMatches("نعم،،", 1)
	// spaced consecutive Arabic commas (same as Java token path after whitespace strip)
	assertMatches("نعم ، ،", 1)
	// double period (base rule applies to all langs)
	assertMatches("نعم..", 1)
	assertMatches("This is a test sentence.. Another sentence", 1)
	// triple Arabic commas: Java only fires when commaCount == 2 && next != commaChar;
	// at the 2nd ، next is still ، so no fire; at the 3rd count is 3 ≠ 2 → 0 matches.
	assertMatches("نعم،،،", 0)
}

func TestArabicDoublePunctuationRule_SpanAndSuggestion(t *testing.T) {
	rule := NewArabicDoublePunctuationRule(arabicDoublePunctMessages())

	// Double Arabic comma glued: "نعم،،" → fromPos=3 toPos=5, suggestion "،"
	matches := rule.Match(arAnalyze("نعم،،"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 5, matches[0].GetToPos())
	require.Equal(t, []string{"،"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Two consecutive commas", matches[0].GetMessage())
	// Java new RuleMatch(this, ...) — rule identity is Arabic subclass
	require.Equal(t, rule, matches[0].GetRule())
	require.Equal(t, "ARABIC_DOUBLE_PUNCTUATION",
		matches[0].GetRule().(interface{ GetID() string }).GetID())

	// Spaced: "نعم ، ،" — second ، at startPos=6 → fromPos=5 toPos=7
	matches = rule.Match(arAnalyze("نعم ، ،"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 5, matches[0].GetFromPos())
	require.Equal(t, 7, matches[0].GetToPos())
	require.Equal(t, []string{"،"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, rule, matches[0].GetRule())

	// Double period: "نعم.." → fromPos=3 toPos=5, suggestions "." and "…"
	matches = rule.Match(arAnalyze("نعم.."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 5, matches[0].GetToPos())
	require.Equal(t, []string{".", "…"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Two consecutive dots", matches[0].GetMessage())
	require.Equal(t, rule, matches[0].GetRule())
}

func TestArabicDoublePunctuationRule_NotDefaultOff(t *testing.T) {
	// Java ArabicDoublePunctuationRule ctor is super(messages) only — no setDefaultOff.
	// Registration must not mark ARABIC_DOUBLE_PUNCTUATION default-off.
	lt := languagetool.NewJLanguageTool("ar")
	RegisterCoreArabicRules(lt)
	for _, id := range lt.GetDefaultOffRuleIDs() {
		require.NotEqual(t, "ARABIC_DOUBLE_PUNCTUATION", id)
	}
}
