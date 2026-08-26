package en

// Outcome twin of org.languagetool.rules.en.SimpleReplaceRule
// (no dedicated Java unit test; fields + AbstractSimpleReplaceRule2 match outcomes
// mirror AmericanReplaceRuleTest / EnglishDiacriticsRuleTest style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(en).getAnalyzedSentence for this rule’s surface matching.
// SimpleReplaceRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official replace.txt
// multiword keys (e.g. "conversation rate") tokenize the same under WordTokenizer as EN.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Match outcomes from official replace.txt rows.
func TestSimpleReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleReplaceRule(TestTools.getMessages("en"), new English())
	//       super(messages, language); useSubRuleSpecificIds(); …
	rule := NewSimpleReplaceRule(nil)

	// correct sentences (already-correct form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Please discourage that behavior."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("We track conversion rate carefully."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nothing to replace here."))))

	// Official replace.txt rows
	checkEnglishSimpleReplace(t, rule, "Please disencourage that behavior.", "discourage")
	checkEnglishSimpleReplace(t, rule, "We track conversation rate carefully.", "conversion rate")
	checkEnglishSimpleReplace(t, rule, "Barbara Streisand sang.", "Barbra Streisand")
	checkEnglishSimpleReplace(t, rule, "A kickass guitar solo", "kick-ass")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	matches := rule.Match(languagetool.AnalyzePlain("Please disencourage that behavior."))
	require.Len(t, matches, 1)
	require.Equal(t, "Wrong word", matches[0].GetShortMessage())
	require.Equal(t,
		"Did you mean <suggestion>discourage</suggestion>?",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) — LanguageCode is getLocale "EN"
	// (toId only special-cases "de"; EN vs en yield the same id string).
	wantID := tools.ToId(EnglishSimpleReplaceRuleID+"_"+"disencourage", "EN")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "EN_SIMPLE_REPLACE_DISENCOURAGE", sr.GetID())
	require.Equal(t, "Check for wrong words/phrases: disencourage", sr.GetDescription())
	require.Equal(t, 7, matches[0].GetFromPos())
	require.Equal(t, 19, matches[0].GetToPos())

	// Multiword sub-rule id (conversation rate)
	matches = rule.Match(languagetool.AnalyzePlain("We track conversation rate carefully."))
	require.Len(t, matches, 1)
	wantID = tools.ToId(EnglishSimpleReplaceRuleID+"_"+"conversation rate", "EN")
	sr, ok = matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected for multiword")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "EN_SIMPLE_REPLACE_CONVERSATION_RATE", sr.GetID())
	require.Equal(t, "Check for wrong words/phrases: conversation rate", sr.GetDescription())
	require.Equal(t, 9, matches[0].GetFromPos())
	require.Equal(t, 26, matches[0].GetToPos())
	require.Equal(t, []string{"conversion rate"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Wrong word", matches[0].GetShortMessage())
	require.Equal(t,
		"Did you mean <suggestion>conversion rate</suggestion>?",
		matches[0].GetMessage())
}

// checkEnglishSimpleReplace asserts one match with the given first replacement.
func checkEnglishSimpleReplace(t *testing.T, rule *SimpleReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of SimpleReplaceRule constructor + overrides.
func TestSimpleReplaceRule_Metadata(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)

	require.Equal(t, EnglishSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "EN_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Check for wrong words/phrases: $match", rule.GetDescription())
	require.Equal(t, "Wrong word", rule.GetShort())
	require.Equal(t, "Did you mean $suggestions?", rule.GetMessage())
	// Java getLocale() → new Locale("EN")
	require.Equal(t, "EN", rule.GetLocaleCode())
	require.Equal(t, EnglishSimpleReplaceFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/en/replace.txt", "/en/replace_custom.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	// Java does not override getSuggestionsSeparator → default ", "
	// (empty Go field maps to ", " in createMatch)
	require.Equal(t, "", rule.SuggestionsSeparator)

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	// Rule default ITSIssueType.Uncategorized
	require.Equal(t, rules.ITSUncategorized, rule.GetLocQualityIssueType())

	// Java SimpleReplaceRule ctor adds no example pairs
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())
}
