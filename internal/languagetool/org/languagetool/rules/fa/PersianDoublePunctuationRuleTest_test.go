package fa

// Outcome twins for org.languagetool.rules.fa.PersianDoublePunctuationRule.
// Java has no dedicated *Test.java for this type; behavior is DoublePunctuationRule
// with getCommaCharacter() = "،" (U+060C) and getId PERSIAN_DOUBLE_PUNCTUATION.
// Upstream DoublePunctuationRuleTest covers base outcomes (dots / ASCII commas);
// here we twin FA-visible paths via PersianWordTokenizer (splits on "،").

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func persianDoublePunctMessages() map[string]string {
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

// faAnalyze is defined in PersianCommaWhitespaceRuleTest_test.go (same package).

func TestPersianDoublePunctuationRule_Metadata(t *testing.T) {
	rule := NewPersianDoublePunctuationRule(persianDoublePunctMessages())
	require.Equal(t, "PERSIAN_DOUBLE_PUNCTUATION", rule.GetID())
	require.Equal(t, "،", rule.GetCommaCharacter())
	require.Equal(t, "Use of two consecutive dots or commas", rule.GetDescription())
}

func TestPersianDoublePunctuationRule_Rule(t *testing.T) {
	rule := NewPersianDoublePunctuationRule(persianDoublePunctMessages())
	assertMatches := func(text string, n int) {
		t.Helper()
		require.Equal(t, n, len(rule.Match(faAnalyze(text))), "text=%q", text)
	}

	// correct — single punctuation / ellipsis / non-Persian comma
	assertMatches("بله،", 0)
	assertMatches("بله، تست", 0)
	assertMatches("این جمله آزمایشی.", 0)
	// three dots (base rule: not exactly two consecutive dots)
	assertMatches("بله...", 0)
	assertMatches("This is a test sentence...", 0)
	// ASCII comma is not getCommaCharacter() → no double-comma match
	assertMatches("بله,,", 0)
	assertMatches("This,, is a test.", 0)
	// paths / Russian ?.. !.. (base DoublePunctuationRuleTest)
	assertMatches("The path is ../data/vtest.avi", 0)
	assertMatches("Это тестовое предложение?..", 0)
	assertMatches("Это тестовое предложение!.. ", 0)
	// single spaced Arabic commas are not consecutive tokens of comma only —
	// "بله ، ،" has tokens "،" "،" with no intervening non-comma token → match (error)
	// handled below under errors.

	// errors — double Arabic comma (U+060C), glued (PersianWordTokenizer splits each ،)
	assertMatches("بله،،", 1)
	// spaced consecutive Arabic commas (same as Java token path after whitespace strip)
	assertMatches("بله ، ،", 1)
	// double period (base rule applies to all langs)
	assertMatches("بله..", 1)
	assertMatches("This is a test sentence.. Another sentence", 1)
	// triple Arabic commas: Java only fires when commaCount == 2 && next != commaChar;
	// at the 2nd ، next is still ، so no fire; at the 3rd count is 3 ≠ 2 → 0 matches.
	assertMatches("بله،،،", 0)
}

func TestPersianDoublePunctuationRule_SpanAndSuggestion(t *testing.T) {
	rule := NewPersianDoublePunctuationRule(persianDoublePunctMessages())

	// Double Arabic comma glued: "بله،،" → fromPos=3 toPos=5, suggestion "،"
	matches := rule.Match(faAnalyze("بله،،"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 5, matches[0].GetToPos())
	require.Equal(t, []string{"،"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Two consecutive commas", matches[0].GetMessage())
	// Java new RuleMatch(this, ...) — rule identity is Persian subclass
	require.Equal(t, rule, matches[0].GetRule())
	require.Equal(t, "PERSIAN_DOUBLE_PUNCTUATION",
		matches[0].GetRule().(interface{ GetID() string }).GetID())

	// Spaced: "بله ، ،" — second ، at startPos=6 → fromPos=5 toPos=7
	matches = rule.Match(faAnalyze("بله ، ،"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 5, matches[0].GetFromPos())
	require.Equal(t, 7, matches[0].GetToPos())
	require.Equal(t, []string{"،"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, rule, matches[0].GetRule())

	// Double period: "بله.." → fromPos=3 toPos=5, suggestions "." and "…"
	matches = rule.Match(faAnalyze("بله.."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 5, matches[0].GetToPos())
	require.Equal(t, []string{".", "…"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Two consecutive dots", matches[0].GetMessage())
	require.Equal(t, rule, matches[0].GetRule())
}

func TestPersianDoublePunctuationRule_NotDefaultOff(t *testing.T) {
	// Java PersianDoublePunctuationRule ctor is super(messages) only — no setDefaultOff.
	// Registration must not mark PERSIAN_DOUBLE_PUNCTUATION default-off.
	lt := languagetool.NewJLanguageTool("fa")
	RegisterCorePersianRules(lt)
	for _, id := range lt.GetDefaultOffRuleIDs() {
		require.NotEqual(t, "PERSIAN_DOUBLE_PUNCTUATION", id)
	}
	// Contrast: PersianCommaWhitespaceRule is default-off in Java/Go registration.
	require.Contains(t, lt.GetDefaultOffRuleIDs(), "PERSIAN_COMMA_PARENTHESIS_WHITESPACE")
}