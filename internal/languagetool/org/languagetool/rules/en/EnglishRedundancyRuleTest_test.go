package en

// Outcome twin of org.languagetool.rules.en.EnglishRedundancyRule
// (no dedicated Java unit test; fields + AbstractSimpleReplaceRule2 match outcomes
// mirror AmericanReplaceRuleTest / EnglishDiacriticsRuleTest / EnglishPlainEnglishRuleTest style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(en).getAnalyzedSentence for this rule’s surface matching.
// EnglishRedundancyRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official redundancies.txt
// multiword keys (e.g. "tuna fish") tokenize the same under WordTokenizer as EN.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official redundancies.txt rows.
func TestEnglishRedundancyRule_Rule(t *testing.T) {
	// Java: new EnglishRedundancyRule(TestTools.getMessages("en"))
	//       super(messages, new English()); setDefaultOff(); … (no useSubRuleSpecificIds)
	rule := NewEnglishRedundancyRule(nil)

	// correct sentences (already non-redundant form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("I ate tuna yesterday."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nothing to replace here."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("The bonus was welcome."))))

	// Java example: tuna fish → tuna
	checkEnglishRedundancyReplace(t, rule, "I ate tuna fish yesterday.", "tuna")

	// Official redundancies.txt rows
	checkEnglishRedundancyReplace(t, rule, "An added bonus for everyone.", "bonus")
	checkEnglishRedundancyReplace(t, rule, "They need advance planning.", "planning")
	checkEnglishRedundancyReplace(t, rule, "The end result was clear.", "result")
	checkEnglishRedundancyReplace(t, rule, "That is a true fact.", "fact")

	// Java-visible createMatch extras (no useSubRuleSpecificIds; message/short templates)
	matches := rule.Match(languagetool.AnalyzePlain("I ate tuna fish yesterday."))
	require.Len(t, matches, 1)
	require.Equal(t, "Redundancy", matches[0].GetShortMessage())
	require.Equal(t,
		"'tuna fish' is a redundancy. In some cases, it might be preferable to use <suggestion>tuna</suggestion>",
		matches[0].GetMessage())
	// Java does not call useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2, not SpecificIdRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, EnglishRedundancyReplaceID, base.GetID())
	// "I ate tuna fish yesterday." → "tuna fish" starts at index 6 (after "I ate ")
	require.Equal(t, 6, matches[0].GetFromPos())
	require.Equal(t, 15, matches[0].GetToPos())
	require.Equal(t, []string{"tuna"}, matches[0].GetSuggestedReplacements())

	// Positions for multiword "added bonus" in "An added bonus for everyone."
	matches = rule.Match(languagetool.AnalyzePlain("An added bonus for everyone."))
	require.Len(t, matches, 1)
	base, ok = matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok)
	require.Equal(t, EnglishRedundancyReplaceID, base.GetID())
	require.Equal(t, "Redundancy", matches[0].GetShortMessage())
	require.Equal(t,
		"'added bonus' is a redundancy. In some cases, it might be preferable to use <suggestion>bonus</suggestion>",
		matches[0].GetMessage())
	// "An " = 3 chars; "added bonus" = 11 chars → 3..14
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 14, matches[0].GetToPos())
	require.Equal(t, []string{"bonus"}, matches[0].GetSuggestedReplacements())
}

// checkEnglishRedundancyReplace asserts one match with the given first replacement.
func checkEnglishRedundancyReplace(t *testing.T, rule *EnglishRedundancyRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of EnglishRedundancyRule constructor + overrides.
func TestEnglishRedundancyRule_Metadata(t *testing.T) {
	rule := NewEnglishRedundancyRule(nil)

	require.Equal(t, EnglishRedundancyReplaceID, rule.GetID())
	require.Equal(t, "EN_REDUNDANCY_REPLACE", rule.GetID())
	require.Equal(t, "1. Redundancy (General)", rule.GetDescription())
	require.Equal(t, "Redundancy", rule.GetShort())
	require.Equal(t,
		"'$match' is a redundancy. In some cases, it might be preferable to use $suggestions",
		rule.GetMessage())
	require.Equal(t, " or ", rule.GetSuggestionsSeparator())
	require.Equal(t, "en", rule.GetLocaleCode())
	require.Equal(t, []string{EnglishRedundancyFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/en/redundancies.txt"}, rule.GetFileNames())
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, EnglishRedundancyURL, rule.GetURL())
	require.Equal(t,
		"https://en.wikipedia.org/wiki/Redundancy_(linguistics)",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "REDUNDANCY", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>tuna fish</marker>", inc[0].GetExample())
	require.Equal(t, []string{"tuna"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>tuna</marker>", corr[0].GetExample())
}
