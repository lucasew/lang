package km

// Outcome twin matrix for org.languagetool.rules.km.KhmerUnpairedBracketsRule
// (no dedicated Java *Test; bar is EN thin-wrapper twin style + GenericUnpairedBracketsRule).
//
// Token path: AnalyzePlain — symbol tokens and sentence-end flags match GenericUnpairedBracketsRule
// expectations for ASCII punctuation / curly quotes / guillemets (same as RU twin convention).
// Unbalanced open is flagged only when a sentence-ending form remains after fillSymbolStack
// (e.g. "(ok." with period), not bare "(ok" — same as EN/core GenericUnpairedBracketsRule.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func kmMessagesUBR() map[string]string {
	// MessagesBundle_en keys used by GenericUnpairedBracketsRule / category.
	return map[string]string{
		"unpaired_brackets":      "Unpaired symbol: ''{0}'' seems to be missing",
		"desc_unpaired_brackets": "Unpaired braces, brackets, quotation marks and similar symbols",
		"category_punctuation":   "Punctuation",
	}
}

func matchNKmUBR(rule *KhmerUnpairedBracketsRule, s string) int {
	return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
}

func TestKhmerUnpairedBracketsRule_Balanced(t *testing.T) {
	rule := NewKhmerUnpairedBracketsRule(kmMessagesUBR())
	// Each KM pair kind balanced → 0
	for _, s := range []string{
		"(ok)",
		"[ok]",
		"{ok}",
		"\"ok\"",
		"'ok'",
		"“ok”",
		"«ok»",
	} {
		require.Equal(t, 0, matchNKmUBR(rule, s), "balanced %q", s)
	}
}

func TestKhmerUnpairedBracketsRule_UnbalancedOpen(t *testing.T) {
	rule := NewKhmerUnpairedBracketsRule(kmMessagesUBR())
	// Unpaired open with sentence-end punctuation (GenericUnpairedBracketsRule convention)
	for _, s := range []string{
		"(ok.",
		"[ok.",
		"{ok.",
		"\"ok.",
		"'ok.",
		"“ok.",
		"«ok.",
	} {
		require.Equal(t, 1, matchNKmUBR(rule, s), "unbalanced open %q", s)
	}
	// Open without sentence-end punctuation is not reported (same as EN/core)
	require.Equal(t, 0, matchNKmUBR(rule, "(ok"))
}

func TestKhmerUnpairedBracketsRule_UnbalancedClose(t *testing.T) {
	rule := NewKhmerUnpairedBracketsRule(kmMessagesUBR())
	// Unpaired close for each pair kind → 1
	for _, s := range []string{
		"ok)",
		"ok]",
		"ok}",
		"ok\"",
		"ok'",
		"ok”",
		"ok»",
	} {
		require.Equal(t, 1, matchNKmUBR(rule, s), "unbalanced close %q", s)
	}
}

// Java KhmerUnpairedBracketsRule: KM_UNPAIRED_BRACKETS, KM symbols, category from base.
func TestKhmerUnpairedBracketsRule_Metadata(t *testing.T) {
	rule := NewKhmerUnpairedBracketsRule(kmMessagesUBR())
	require.Equal(t, "KM_UNPAIRED_BRACKETS", rule.GetID())
	require.Equal(t, []string{"[", "(", "{", "“", "\"", "'", "«"}, rule.StartSymbols)
	require.Equal(t, []string{"]", ")", "}", "”", "\"", "'", "»"}, rule.EndSymbols)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "PUNCTUATION", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSTypographical, rule.GetLocQualityIssueType())
	require.Equal(t, kmMessagesUBR()["desc_unpaired_brackets"], rule.GetDescription())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	// MatchList rebinds Rule so match.Rule is the KM type with KM id
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("(ok."),
	})
	require.NotEmpty(t, ms)
	require.Equal(t, rule, ms[0].Rule)
	require.Equal(t, "KM_UNPAIRED_BRACKETS", rule.GetID())
}

// Keep a single smoke name for historical discoverability.
func TestKhmerUnpairedBracketsRule(t *testing.T) {
	rule := NewKhmerUnpairedBracketsRule(kmMessagesUBR())
	require.Equal(t, "KM_UNPAIRED_BRACKETS", rule.GetID())
	require.Equal(t, 0, matchNKmUBR(rule, "(ok)"))
	require.Equal(t, 1, matchNKmUBR(rule, "(ok."))
	require.Equal(t, 1, matchNKmUBR(rule, "ok)"))
}
