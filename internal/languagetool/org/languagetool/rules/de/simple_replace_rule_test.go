package de

// Outcome twin of org.languagetool.rules.de.SimpleReplaceRule
// (no dedicated Java unit test; fields + AbstractSimpleReplaceRule2 match outcomes
// mirror EN SimpleReplaceRuleTest / AmericanReplaceRuleTest style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(de).getAnalyzedSentence for this rule’s surface matching.
// SimpleReplaceRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official replace.txt
// multiword keys (e.g. "Conversation Rate", "Jo Biden") tokenize the same under
// WordTokenizer as DE.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Match outcomes from official replace.txt rows.
func TestSimpleReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleReplaceRule(TestTools.getMessages("de"), new German())
	//       super(messages, language); useSubRuleSpecificIds(); …
	rule := NewSimpleReplaceRule(nil)

	// correct sentences (already-correct form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Joe Biden sprach gestern."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Wir messen Conversion Rate sorgfältig."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nichts zu ersetzen hier."))))

	// Official replace.txt rows
	checkGermanSimpleReplace(t, rule, "Jo Biden sprach gestern.", "Joe Biden")
	checkGermanSimpleReplace(t, rule, "Wir messen Conversation Rate sorgfältig.", "Conversion Rate")
	checkGermanSimpleReplace(t, rule, "Nicholas Cage spielte mit.", "Nicolas Cage")
	checkGermanSimpleReplace(t, rule, "Jo Bidens Rede war kurz.", "Joe Bidens")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	matches := rule.Match(languagetool.AnalyzePlain("Jo Biden sprach gestern."))
	require.Len(t, matches, 1)
	require.Equal(t, "Falsches Wort", matches[0].GetShortMessage())
	require.Equal(t,
		"Meinten Sie vielleicht <suggestion>Joe Biden</suggestion>?",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) — LanguageCode is getLocale "DE"
	// (toId only special-cases "de"; DE vs de differ on umlauts only; EN twin uses Locale form).
	wantID := tools.ToId(GermanSimpleReplaceRuleID+"_"+"Jo Biden", "DE")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "DE_SIMPLE_REPLACE_JO_BIDEN", sr.GetID())
	require.Equal(t, "Prüft auf bestimmte falsche Wörter/Phrasen: Jo Biden", sr.GetDescription())
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 8, matches[0].GetToPos())

	// Multiword sub-rule id (Conversation Rate)
	matches = rule.Match(languagetool.AnalyzePlain("Wir messen Conversation Rate sorgfältig."))
	require.Len(t, matches, 1)
	wantID = tools.ToId(GermanSimpleReplaceRuleID+"_"+"Conversation Rate", "DE")
	sr, ok = matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected for multiword")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "DE_SIMPLE_REPLACE_CONVERSATION_RATE", sr.GetID())
	require.Equal(t, "Prüft auf bestimmte falsche Wörter/Phrasen: Conversation Rate", sr.GetDescription())
	require.Equal(t, 11, matches[0].GetFromPos())
	require.Equal(t, 28, matches[0].GetToPos())
	require.Equal(t, []string{"Conversion Rate"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Falsches Wort", matches[0].GetShortMessage())
	require.Equal(t,
		"Meinten Sie vielleicht <suggestion>Conversion Rate</suggestion>?",
		matches[0].GetMessage())
}

// checkGermanSimpleReplace asserts one match with the given first replacement.
func checkGermanSimpleReplace(t *testing.T, rule *SimpleReplaceRule, sentence, word string) {
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

	require.Equal(t, GermanSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "DE_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Prüft auf bestimmte falsche Wörter/Phrasen: $match", rule.GetDescription())
	require.Equal(t, "Falsches Wort", rule.GetShort())
	require.Equal(t, "Meinten Sie vielleicht $suggestions?", rule.GetMessage())
	// Java getLocale() → new Locale("DE")
	require.Equal(t, "DE", rule.GetLocaleCode())
	require.Equal(t, GermanSimpleReplaceFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/de/replace.txt", "/de/replace_custom.txt"}, rule.GetFileNames())
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
