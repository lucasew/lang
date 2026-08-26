package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/SimpleReplaceRuleTest.java
//
// Java setUp:
//   lt = new JLanguageTool(new Polish());
//   rule = new SimpleReplaceRule(TestTools.getMessages("pl"), lt.getLanguage());
// Java testRule → rule.match(lt.getAnalyzedSentence(...))
// Analysis ports JLanguageTool(Polish).getAnalyzedSentence — PL tokenizer +
// PolishTagger + hybrid disambiguator (same as analyzePLSpeller / CompoundRule).
// Required for "Witamy prez. Komorowskiego!" — disambiguation.xml rule id "prez"
// applies ignore_spelling to "prez" + "." (ASR Match skips IsIgnoredBySpeller).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Port of SimpleReplaceRuleTest.testRule
func TestSimpleReplaceRule_Rule(t *testing.T) {
	// Java: lt = new JLanguageTool(new Polish()); rule = new SimpleReplaceRule(...)
	rule := NewSimpleReplaceRule(nil)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(analyzePLSpeller(t, "Wszystko w porządku."))))

	// no checking lemmas:
	require.Equal(t, 0, len(rule.Match(analyzePLSpeller(t, "Pola lodowe"))))

	// with immunized tokens (Java comment; disambig action is ignore_spelling on "prez."):
	require.Equal(t, 0, len(rule.Match(analyzePLSpeller(t, "Witamy prez. Komorowskiego!"))))

	// incorrect sentences:
	// at the beginning of a sentence (Romanian replace rule is case-sensitive)
	checkSimpleReplaceRule(t, rule, "Piaty przypadek.", "Piąty")
	// inside sentence
	checkSimpleReplaceRule(t, rule, "To piaty przypadek.", "piąty")
}

// checkSimpleReplaceRule ports SimpleReplaceRuleTest.checkSimpleReplaceRule.
func checkSimpleReplaceRule(t *testing.T, rule *SimpleReplaceRule, sentence, word string) {
	t.Helper()
	// Java: rule.match(lt.getAnalyzedSentence(sentence))
	matches := rule.Match(analyzePLSpeller(t, sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count wile checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of SimpleReplaceRule constructor + overrides.
func TestSimpleReplaceRule_Metadata(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)

	require.Equal(t, PolishSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "PL_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Typowe literówki i niepoprawne wyrazy (domowi, sie, niewiadomo, duh, cie…)", rule.GetDescription())
	require.Equal(t, "Literówka", rule.GetShort())
	// MessageFn with real args (Java getMessage)
	require.Equal(t,
		"Wyraz „piaty” to najczęściej literówka; poprawnie pisze się: piąty.",
		rule.MessageFn("piaty", []string{"piąty"}))
	require.Equal(t,
		"Wyraz „x” to najczęściej literówka; poprawnie pisze się: a, b.",
		rule.MessageFn("x", []string{"a", "b"}))

	// Java getLocale() → Locale("pl")
	require.Equal(t, "pl", rule.GetLocaleCode())
	require.Equal(t, "pl", rule.LanguageCode)
	require.False(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	require.False(t, rule.SubRuleSpecificIDs)
	require.False(t, rule.IgnoreTaggedWords)
	require.Equal(t, "", rule.GetURL())
	require.False(t, rule.IsDefaultOff())

	// Java setCategory + setLocQualityIssueType
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "PRAWDOPODOBNE_LITEROWKI", rule.GetCategory().GetID().String())
	require.Equal(t, "Prawdopodobne literówki", rule.GetCategory().GetName())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Java example pair: sei → się
	require.Equal(t, []string{"się"}, rule.GetIncorrectExamples()[0].GetCorrections())

	// Official /pl/replace.txt rows
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"piąty"}, ww["piaty"])
	require.Equal(t, []string{"się"}, ww["sei"])
	require.Equal(t, []string{"się"}, ww["sie"])
	require.Equal(t, []string{"przez"}, ww["prez"])

	// Match-time message/short for piaty hit
	matches := rule.Match(analyzePLSpeller(t, "To piaty przypadek."))
	require.Len(t, matches, 1)
	require.Equal(t, "Literówka", matches[0].GetShortMessage())
	require.Equal(t,
		"Wyraz „piaty” to najczęściej literówka; poprawnie pisze się: piąty.",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, PolishSimpleReplaceRuleID, base.GetID())
	require.Equal(t, "PRAWDOPODOBNE_LITEROWKI", base.GetCategory().GetID().String())
}
