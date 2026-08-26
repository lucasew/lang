package be

// Twin of languagetool-language-modules/be/src/test/java/org/languagetool/rules/be/SimpleReplaceRuleTest.java
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(be).getAnalyzedSentence for this rule’s surface matching.
// SimpleReplaceRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official replace.txt
// keys tokenize the same under WordTokenizer as BE (accepted sibling pattern:
// RussianSimpleReplaceRule / EN SimpleReplaceRule).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java SimpleReplaceRuleTest.testRule full matrix.
func TestSimpleReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleReplaceRule(TestTools.getMessages("be"));
	//       JLanguageTool lt = new JLanguageTool(new Belarusian());
	rule := NewSimpleReplaceRule(nil)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Напрамкі дзейнасці былі ўзгоднены з ,"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Яго камп’ютар выключыўся."))))

	// incorrect sentences:
	// Java: suggestion "камп’ютар" uses U+2019 RIGHT SINGLE QUOTATION MARK (official replace.txt)
	matches := rule.Match(languagetool.AnalyzePlain("Яго кампутар выключыўся."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	// "камп" + U+2019 + "ютар" — must match Java resource codepoints, not ASCII apostrophe
	wantSugg := "камп\u2019ютар"
	require.Equal(t, wantSugg, matches[0].GetSuggestedReplacements()[0])
	// Java-visible createMatch short / message templates (no sub-rule ids)
	require.Equal(t, "Памылка?", matches[0].GetShortMessage())
	require.Equal(t,
		"«кампутар» — памылка, нелітаратурны выраз або прастамоўе, правільна: <suggestion>"+wantSugg+"</suggestion>",
		matches[0].GetMessage())
	// Java does not call useSubRuleSpecificIds() → match rule is the parent rule, not SpecificIdRule
	_, isSub := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSub, "no sub-rule specific id expected")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok)
	require.Equal(t, BelarusianSimpleReplaceRuleID, base.GetID())
}

// Field twin of SimpleReplaceRule constructor + overrides.
func TestSimpleReplaceRule_Metadata(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)

	require.Equal(t, BelarusianSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "BE_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Пошук прастамоўяў і памылковых фраз", rule.GetDescription())
	require.Equal(t, "Памылка?", rule.GetShort())
	require.Equal(t, "«$match» — памылка, нелітаратурны выраз або прастамоўе, правільна: $suggestions", rule.GetMessage())
	// Java getLocale() → new Locale("be")
	require.Equal(t, "be", rule.GetLocaleCode())
	require.Equal(t, BelarusianSimpleReplaceFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/be/replace.txt"}, rule.GetFileNames())
	// Java getSuggestionsSeparator override
	require.Equal(t, ", ", rule.GetSuggestionsSeparator())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	// Java: setCategory(new Category(new CategoryId("MISC"), "Агульныя правілы"))
	require.Equal(t, "Агульныя правілы", rule.GetCategory().GetName())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t,
		"<marker>З большага</marker>, гэта быў добры дзень.",
		inc[0].GetExample())
	require.Equal(t, []string{"Збольшага"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t,
		"<marker>Збольшага</marker>, гэта быў добры дзень.",
		corr[0].GetExample())
}
