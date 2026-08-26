package en

// Outcome twin of org.languagetool.rules.en.EnglishPlainEnglishRule
// (no dedicated Java unit test; fields + AbstractSimpleReplaceRule2 match outcomes
// mirror AmericanReplaceRuleTest / EnglishDiacriticsRuleTest style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(en).getAnalyzedSentence for this rule’s surface matching.
// EnglishPlainEnglishRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official wordiness.txt
// multiword keys (e.g. "fatal outcome") tokenize the same under WordTokenizer as EN.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official wordiness.txt rows.
func TestEnglishPlainEnglishRule_Rule(t *testing.T) {
	// Java: new EnglishPlainEnglishRule(TestTools.getMessages("en"))
	//       super(messages, new English()); setDefaultOff(); … (no useSubRuleSpecificIds)
	rule := NewEnglishPlainEnglishRule(nil)

	// correct sentences (plain form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("A death occurred."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nothing to replace here."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("We will act if needed."))))

	// Java example: fatal outcome → death
	checkEnglishPlainEnglishReplace(t, rule, "A fatal outcome occurred.", "death")

	// Official wordiness.txt rows
	checkEnglishPlainEnglishReplace(t, rule, "I myself will go.", "I")
	checkEnglishPlainEnglishReplace(t, rule, "We will act at this point in time.", "now")
	checkEnglishPlainEnglishReplace(t, rule, "Call me in the event that you need help.", "if")

	// Multi-suggestion row: due to the fact that → due to|since
	matches := rule.Match(languagetool.AnalyzePlain("We left due to the fact that it rained."))
	require.Len(t, matches, 1)
	require.Equal(t, []string{"due to", "since"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Wordiness", matches[0].GetShortMessage())
	require.Equal(t,
		"'due to the fact that' is a wordy or complex expression. In some cases, it might be preferable to use <suggestion>due to</suggestion> or <suggestion>since</suggestion>.",
		matches[0].GetMessage())
	// Java does not call useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule2, not SpecificIdRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok, "base rule expected without sub-rule ids")
	require.Equal(t, EnglishPlainEnglishReplaceID, base.GetID())
	require.Equal(t, 8, matches[0].GetFromPos())
	require.Equal(t, 28, matches[0].GetToPos())

	// Java example positions: "A fatal outcome occurred." → "fatal outcome" at 2..15
	matches = rule.Match(languagetool.AnalyzePlain("A fatal outcome occurred."))
	require.Len(t, matches, 1)
	base, ok = matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok)
	require.Equal(t, EnglishPlainEnglishReplaceID, base.GetID())
	require.Equal(t, "Wordiness", matches[0].GetShortMessage())
	require.Equal(t,
		"'fatal outcome' is a wordy or complex expression. In some cases, it might be preferable to use <suggestion>death</suggestion>.",
		matches[0].GetMessage())
	require.Equal(t, 2, matches[0].GetFromPos())
	require.Equal(t, 15, matches[0].GetToPos())
	require.Equal(t, []string{"death"}, matches[0].GetSuggestedReplacements())
}

// checkEnglishPlainEnglishReplace asserts one match with the given first replacement.
func checkEnglishPlainEnglishReplace(t *testing.T, rule *EnglishPlainEnglishRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of EnglishPlainEnglishRule constructor + overrides.
func TestEnglishPlainEnglishRule_Metadata(t *testing.T) {
	rule := NewEnglishPlainEnglishRule(nil)

	require.Equal(t, EnglishPlainEnglishReplaceID, rule.GetID())
	require.Equal(t, "EN_PLAIN_ENGLISH_REPLACE", rule.GetID())
	require.Equal(t, "1. Wordiness (General)", rule.GetDescription())
	require.Equal(t, "Wordiness", rule.GetShort())
	require.Equal(t,
		"'$match' is a wordy or complex expression. In some cases, it might be preferable to use $suggestions.",
		rule.GetMessage())
	require.Equal(t, " or ", rule.GetSuggestionsSeparator())
	require.Equal(t, "en", rule.GetLocaleCode())
	require.Equal(t, []string{EnglishPlainEnglishFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/en/wordiness.txt"}, rule.GetFileNames())
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, EnglishPlainEnglishURL, rule.GetURL())
	require.Equal(t,
		"https://en.wikipedia.org/wiki/List_of_plain_English_words_and_phrases",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "PLAIN_ENGLISH", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>fatal outcome</marker>", inc[0].GetExample())
	require.Equal(t, []string{"death"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>death</marker>", corr[0].GetExample())
}
