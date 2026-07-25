package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanParagraphRepeatBeginningRuleTest.java
//
// Java setup:
//   Language lang = Languages.getLanguageForShortCode("de-DE");
//   JLanguageTool lt = new JLanguageTool(lang);
//   setUpRule(lt): disable all rules; add GermanParagraphRepeatBeginningRule; enable
//   assertEquals(N, lt.check(text).size());
//
// Go: de-DE JLanguageTool + WireGermanTagWord (createDefaultTagger) so Analyze/Check
// get real GermanTagger readings — no AnalyzePlain / withART invent for the Java matrix.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// deMessagesParaRepeat ports TestTools.getMessages("de") keys used by ParagraphRepeatBeginningRule.
func deMessagesParaRepeat() map[string]string {
	return map[string]string{
		"repetition_paragraph_beginning_desc":     "Gleicher Anfang von aufeinanderfolgenden Absätzen",
		"repetition_paragraph_beginning_last_msg": "Der Absatz beginnt so wie der vorhergehende Absatz",
		"category_style": "Stil",
	}
}

// setUpGermanParagraphRepeatBeginningRule ports GermanParagraphRepeatBeginningRuleTest.setUpRule.
func setUpGermanParagraphRepeatBeginningRule(t *testing.T) (*GermanParagraphRepeatBeginningRule, *languagetool.JLanguageTool) {
	t.Helper()
	// Java: Languages.getLanguageForShortCode("de-DE")
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for GermanParagraphRepeatBeginningRule twin")
	// Java: for (Rule rule : lt.getAllRules()) { lt.disableRule(rule.getId()); }
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		lt.DisableRule(id)
	}
	// Java: new GermanParagraphRepeatBeginningRule(TestTools.getMessages(lang.getShortCode()), lang)
	rule := NewGermanParagraphRepeatBeginningRule(deMessagesParaRepeat())
	lt.AddTextLevelRuleChecker(rule.GetID(), rules.AsTextLevelChecker(rule.MatchList))
	lt.EnableRule(rule.GetID())
	return rule, lt
}

// TestGermanParagraphRepeatBeginningRule_Rule ports GermanParagraphRepeatBeginningRuleTest.testRule.
func TestGermanParagraphRepeatBeginningRule_Rule(t *testing.T) {
	rule, lt := setUpGermanParagraphRepeatBeginningRule(t)

	// Java subclass meta (getId + inherited STYLE / defaultOff / minToCheckParagraph).
	require.Equal(t, "GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE", rule.GetID())
	require.Equal(t, "Gleicher Anfang von aufeinanderfolgenden Absätzen", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, 1, rule.MinToCheckParagraph())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.NewCategoryId("STYLE")))
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	// Java assertEquals(N, lt.check(...).size()) — full matrix:
	require.Equal(t, 2, len(lt.Check("Der Hund spazierte über die Straße.\n\nDer Hund ignorierte den Verkehr.")))
	require.Equal(t, 0, len(lt.Check("Der Hund spazierte über die Straße.\n\nDas Tier ignorierte den Verkehr.")))
	require.Equal(t, 2, len(lt.Check("Peter spazierte über die Straße.\n\nPeter ignorierte den Verkehr.")))
	require.Equal(t, 0, len(lt.Check("Peter spazierte über die Straße.\n\nDer Junge ignorierte den Verkehr.")))
	require.Equal(t, 2, len(lt.Check("»Peter spazierte über die Straße.«\n\n»Peter ignorierte den Verkehr.«")))
}

// TestGermanParagraphRepeatBeginningRule_IsArticle ports Java isArticle ART POS override
// (GermanParagraphRepeatBeginningRule.isArticle → hasPosTagStartingWith("ART")).
// Real GermanTagger.Tag (ignoreCase sentence path) supplies ART for "Der".
func TestGermanParagraphRepeatBeginningRule_IsArticle(t *testing.T) {
	rule := NewGermanParagraphRepeatBeginningRule(deMessagesParaRepeat())
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger, "GermanTagger resources required")

	// Java createDefaultTagger().tag: "Der" → ART:* (sentence-start lowercase merge).
	der := tagger.Tag([]string{"Der"})[0]
	require.Equal(t, "Der", der.GetToken())
	require.True(t, der.HasPosTagStartingWith("ART"))
	require.True(t, rule.IsArticle(der))

	// Non-article noun is not ART.
	hund := tagger.Tag([]string{"Hund"})[0]
	require.False(t, hund.HasPosTagStartingWith("ART"))
	require.False(t, rule.IsArticle(hund))

	// Untagged token is not an article (Java ART only — no surface invent).
	untagged := languagetool.AnalyzePlain("Xyz").GetTokensWithoutWhitespace()[1]
	require.False(t, rule.IsArticle(untagged))
}

// TestGermanParagraphRepeatBeginningRule_MatchList_RealAnalyze asserts MatchList on
// lt.Analyze (same sentences as Check) — Java-visible match counts without invent.
func TestGermanParagraphRepeatBeginningRule_MatchList_RealAnalyze(t *testing.T) {
	rule, lt := setUpGermanParagraphRepeatBeginningRule(t)

	assertN := func(want int, text string) {
		t.Helper()
		ms := rule.MatchList(lt.Analyze(text))
		require.Equal(t, want, len(ms), "MatchList %q", text)
	}
	assertN(2, "Der Hund spazierte über die Straße.\n\nDer Hund ignorierte den Verkehr.")
	assertN(0, "Der Hund spazierte über die Straße.\n\nDas Tier ignorierte den Verkehr.")
	assertN(2, "Peter spazierte über die Straße.\n\nPeter ignorierte den Verkehr.")
	assertN(0, "Peter spazierte über die Straße.\n\nDer Junge ignorierte den Verkehr.")
	assertN(2, "»Peter spazierte über die Straße.«\n\n»Peter ignorierte den Verkehr.«")
}
