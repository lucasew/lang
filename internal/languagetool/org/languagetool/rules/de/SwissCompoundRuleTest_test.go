package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/SwissCompoundRuleTest.java
//
// Java:
//   lt = new JLanguageTool(Languages.getLanguageForShortCode("de-DE"));
//   rule = new SwissCompoundRule(...);
//   //check(1, "Ente süss-sauer");
//   check(1, "Ente süß-sauer");
//   //check(1, "Ente süss sauer");
//   check(1, "Ente süß sauer");
//
// Shared GermanCompoundRuleTest.runTests matrix also runs Swiss via TestGermanCompoundRule_Rule.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSwissCompoundRule_Rule(t *testing.T) {
	// Java SwissCompoundRuleTest.testRule
	lt := languagetool.NewJLanguageTool("de-DE")
	WireGermanRuntimeResourcesFor(lt, "DE")
	rule := NewSwissCompoundRule(nil)
	require.Equal(t, "DE_CH_COMPOUNDS", rule.GetID())

	check := func(expected int, text string) {
		t.Helper()
		// Java AbstractCompoundRuleTest.check → lt.getAnalyzedSentence
		sent := lt.GetAnalyzedSentence(text)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", text)
		ms := rule.Match(sent)
		require.Equal(t, expected, len(ms), "text %q", text)
	}
	// Java active lines only (commented süss forms are not Java-required):
	//check(1, "Ente süss-sauer");
	check(1, "Ente süß-sauer")
	//check(1, "Ente süss sauer");
	check(1, "Ente süß sauer")
}
