package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/SentenceWhitespaceRuleTest.java
//
// Java setup:
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("de-DE"));
//   TestTools.disableAllRulesExcept(lt, "DE_SENTENCE_WHITESPACE");
//   assertGood/assertBad → lt.check(text).size() is 0 or 1
//   messages → lt.check(...).get(0).getMessage().contains(...)
//
// Go: de-DE JLanguageTool + only DE_SENTENCE_WHITESPACE text-level checker
// (equivalent to disableAllRulesExcept when no other rules are registered).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// setupDESentenceWhitespaceLT ports:
//   new JLanguageTool(de-DE) + disableAllRulesExcept("DE_SENTENCE_WHITESPACE")
func setupDESentenceWhitespaceLT() (*SentenceWhitespaceRule, *languagetool.JLanguageTool) {
	lt := languagetool.NewJLanguageTool("de-DE")
	// Java disableAllRulesExcept: disable every active rule, leave only the given id.
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		if id != "DE_SENTENCE_WHITESPACE" {
			lt.DisableRule(id)
		}
	}
	rule := NewSentenceWhitespaceRule(nil)
	// Ensure only this rule is active (fresh LT has none; RegisterCore not required
	// for sentence-boundary analysis — SRX de-DE matches Java Language.getSentenceTokenizer).
	lt.AddTextLevelRuleChecker(rule.GetID(), rules.AsTextLevelChecker(rule.MatchList))
	return rule, lt
}

// assertGoodSWR ports SentenceWhitespaceRuleTest.assertGood → check size 0.
func assertGoodSWR(t *testing.T, lt *languagetool.JLanguageTool, text string) {
	t.Helper()
	ms := lt.Check(text)
	require.Equal(t, 0, len(ms), "assertGood %q got %d matches: %+v", text, len(ms), ms)
}

// assertBadSWR ports SentenceWhitespaceRuleTest.assertBad → check size 1.
func assertBadSWR(t *testing.T, lt *languagetool.JLanguageTool, text string) {
	t.Helper()
	ms := lt.Check(text)
	require.Equal(t, 1, len(ms), "assertBad %q got %d matches: %+v", text, len(ms), ms)
	require.Equal(t, "DE_SENTENCE_WHITESPACE", ms[0].RuleID)
}

// TestSentenceWhitespaceRule_Match ports SentenceWhitespaceRuleTest.testMatch.
func TestSentenceWhitespaceRule_Match(t *testing.T) {
	rule, lt := setupDESentenceWhitespaceLT()

	// Java DE subclass meta (constructor + overrides).
	require.Equal(t, "DE_SENTENCE_WHITESPACE", rule.GetID())
	require.Equal(t, "Fehlendes Leerzeichen zwischen Sätzen oder nach Ordnungszahlen", rule.GetDescription())
	require.Equal(t,
		"https://languagetool.org/insights/de/beitrag/grammatik-leerzeichen/#fehler-1-leerzeichen-vor-und-nach-satzzeichen",
		rule.GetURL(),
	)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("MISC"), rule.GetCategory().GetID())
	require.Equal(t, rules.ITSWhitespace, rule.GetLocQualityIssueType())
	require.NotEmpty(t, rule.GetIncorrectExamples())
	require.Equal(t, "Hier steht ein Satz.<marker>Das</marker> ist ein weiterer Satz.", rule.GetIncorrectExamples()[0].GetExample())
	require.Equal(t, []string{" Das"}, rule.GetIncorrectExamples()[0].GetCorrections())
	require.NotEmpty(t, rule.GetCorrectExamples())
	require.Equal(t, "Hier steht ein Satz.<marker> Das</marker> ist ein weiterer Satz.", rule.GetCorrectExamples()[0].GetExample())

	// Java getMessage(boolean) strings.
	require.Equal(t, "Fügen Sie nach Ordnungszahlen (1., 2. usw.) ein Leerzeichen ein.", rule.GetMessage(true))
	require.Equal(t, "Fügen Sie zwischen Sätzen ein Leerzeichen ein.", rule.GetMessage(false))

	// Java assertGood:
	assertGoodSWR(t, lt, "Das ist ein Satz. Und hier der nächste.")
	assertGoodSWR(t, lt, "Das ist ein Satz! Und hier der nächste.")
	assertGoodSWR(t, lt, "Ist das ein Satz? Hier der nächste.")
	assertGoodSWR(t, lt, "Am 28. September.")
	assertGoodSWR(t, lt, "Das 1. Internationale Filmfestival findet nächste Woche statt.")

	// Java assertBad:
	assertBadSWR(t, lt, "Das ist ein Satz.Und hier der nächste.")
	assertBadSWR(t, lt, "Das ist ein Satz!Und hier der nächste.")
	assertBadSWR(t, lt, "Ist das ein Satz?Hier der nächste.")
	assertBadSWR(t, lt, "Am 28.September.")
	assertBadSWR(t, lt, "Das 1.Internationale Filmfestival findet nächste Woche statt.")

	// Java message content via lt.check:
	// assertTrue(lt.check("Am 7.September 2014.").get(0).getMessage().contains("nach Ordnungszahlen"));
	ms := lt.Check("Am 7.September 2014.")
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].Message, "nach Ordnungszahlen")

	// assertTrue(lt.check("Im September.Dann der nächste Satz.").get(0).getMessage().contains("zwischen Sätzen"));
	ms = lt.Check("Im September.Dann der nächste Satz.")
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].Message, "zwischen Sätzen")
}
