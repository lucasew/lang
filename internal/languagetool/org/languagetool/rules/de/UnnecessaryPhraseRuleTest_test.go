package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UnnecessaryPhraseRuleTest.java
//
// Java:
//   Language lang = Languages.getLanguageForShortCode("de-DE");
//   JLanguageTool lt = new JLanguageTool(lang);
//   setUpRule(lt): disable all; add UnnecessaryPhraseRule(..., new UserConfig()); enable
//   assertEquals(4, lt.check("In diesem Zusammenhang ist es im Allgemeinen voll und ganz zumindest mehr oder weniger sinnvoll.").size());
//   assertEquals(0, lt.check("„In diesem Zusammenhang ist es im Allgemeinen voll und ganz zumindest mehr oder weniger sinnvoll.“").size());
//   assertEquals(0, lt.check("Es ist weniger sinnvoll.").size());
//
// Go: de-DE analysis via GermanWordTokenizer + GermanTagger.Tag (same path as
// NonSignificantVerbsRule twin) so hasLemma("drücken") exception is real, not invent.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/stretchr/testify/require"
)

// setUpUnnecessaryPhraseRule ports UnnecessaryPhraseRuleTest.setUpRule (+ tagger ready).
func setUpUnnecessaryPhraseRule(t *testing.T) (*UnnecessaryPhraseRule, *languagetool.JLanguageTool, *detag.GermanTagger) {
	t.Helper()
	// Java: Languages.getLanguageForShortCode("de-DE")
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for UnnecessaryPhraseRule twin")
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger, "GermanTagger.INSTANCE resources required")

	// Java: for (Rule rule : lt.getAllRules()) { lt.disableRule(rule.getId()); }
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		lt.DisableRule(id)
	}
	// Java: new UnnecessaryPhraseRule(TestTools.getMessages(lang.getShortCode()), lang, new UserConfig())
	rule := NewUnnecessaryPhraseRule(nil)
	lt.AddTextLevelRuleChecker(rule.GetID(), rules.AsTextLevelChecker(rule.MatchList))
	lt.EnableRule(rule.GetID())
	return rule, lt, tagger
}

// uprCheck ports lt.check(text).size() with only UnnecessaryPhraseRule enabled.
func uprCheck(t *testing.T, rule *UnnecessaryPhraseRule, lt *languagetool.JLanguageTool, tagger *detag.GermanTagger, text string) int {
	t.Helper()
	sents := analyzeTextWithGermanTagger(lt, tagger, text)
	return len(rule.MatchList(sents))
}

// TestUnnecessaryPhraseRule_Rule ports UnnecessaryPhraseRuleTest.testRule.
func TestUnnecessaryPhraseRule_Rule(t *testing.T) {
	rule, lt, tagger := setUpUnnecessaryPhraseRule(t)

	// Java subclass surface (id, default limit, description, default-off).
	require.Equal(t, "UNNECESSARY_PHRASES_DE", rule.GetID())
	require.Equal(t, unnecessaryPhraseDefaultMinPerMill, rule.MinPercent)
	require.Equal(t, 8, rule.MinPercent)
	require.Equal(t, float64(10000), rule.Denominator)
	require.True(t, rule.ExcludeDirectSpeech)
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, "Statistische Stilanalyse: Potenzielle Phrasen", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.NewCategoryId("CREATIVE_WRITING")))
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.NotEmpty(t, rule.GetIncorrectExamples())

	// Java: 4 phrases — "In diesem Zusammenhang", "im Allgemeinen", "voll und ganz", "mehr oder weniger"
	// ("zumindest" is not in the phrase list); percent over 8‰ emits all four.
	require.Equal(t, 4, uprCheck(t, rule, lt, tagger,
		"In diesem Zusammenhang ist es im Allgemeinen voll und ganz zumindest mehr oder weniger sinnvoll."))

	// Java: exclude direct speech with German low/high quotes „…“
	require.Equal(t, 0, uprCheck(t, rule, lt, tagger,
		"„In diesem Zusammenhang ist es im Allgemeinen voll und ganz zumindest mehr oder weniger sinnvoll.“"))

	// Java: no full phrase (only "weniger")
	require.Equal(t, 0, uprCheck(t, rule, lt, tagger, "Es ist weniger sinnvoll."))
}

// TestUnnecessaryPhraseRule_ExampleAndException covers constructor example + an+sich+drücken.
func TestUnnecessaryPhraseRule_ExampleAndException(t *testing.T) {
	rule, lt, tagger := setUpUnnecessaryPhraseRule(t)
	// class example pair (limit 8‰; short sentence still exceeds)
	require.Equal(t, 1, uprCheck(t, rule, lt, tagger, "Das ist allem Anschein nach eine Phrase."))
	require.Equal(t, 0, uprCheck(t, rule, lt, tagger, "Das ist eine Phrase."))

	// Java isException: "an"+"sich" skipped when any token hasLemma("drücken")
	rule0 := NewUnnecessaryPhraseRule(nil)
	rule0.MinPercent = 0
	require.Equal(t, 1, len(rule0.MatchList(analyzeTextWithGermanTagger(lt, tagger, "Das ist an sich gut."))))
	require.Equal(t, 0, len(rule0.MatchList(analyzeTextWithGermanTagger(lt, tagger, "Das ist an sich drücken."))))
}

// MinPercent==0 disables direct-speech skip (Java compound guard in AbstractStatisticStyleRule).
func TestUnnecessaryPhraseRule_MinZeroStillSeesDirectSpeech(t *testing.T) {
	rule, lt, tagger := setUpUnnecessaryPhraseRule(t)
	rule.MinPercent = 0
	// With min 0, phrases inside „…“ still fire (unless WithoutDirectSpeech).
	// first "In diesem Zusammenhang" may miss if nToken!=1 (quote is token 1) so capitalization
	// keeps "In" ≠ "in"; remaining three still match.
	ms := rule.MatchList(analyzeTextWithGermanTagger(lt, tagger,
		"„In diesem Zusammenhang ist es im Allgemeinen voll und ganz zumindest mehr oder weniger sinnvoll.“"))
	require.GreaterOrEqual(t, len(ms), 3)
}
