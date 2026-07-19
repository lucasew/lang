package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/SentenceWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSentenceWhitespaceRule_Match(t *testing.T) {
	rule := NewSentenceWhitespaceRule(nil)
	require.Equal(t, "DE_SENTENCE_WHITESPACE", rule.GetID())
	matchN := func(s string) int {
		return len(rule.MatchList(languagetool.SplitAndAnalyze(s)))
	}
	require.Equal(t, 0, matchN("Das ist ein Satz. Und hier der nächste."))
	require.Equal(t, 0, matchN("Das ist ein Satz! Und hier der nächste."))
	require.Equal(t, 0, matchN("Ist das ein Satz? Hier der nächste."))

	require.Equal(t, 1, matchN("Das ist ein Satz.Und hier der nächste."))
	require.Equal(t, 1, matchN("Das ist ein Satz!Und hier der nächste."))
	require.Equal(t, 1, matchN("Ist das ein Satz?Hier der nächste."))

	require.Equal(t, 0, matchN("Am 28. September."))
	require.Equal(t, 1, matchN("Am 28.September."))
	require.Equal(t, "Fehlendes Leerzeichen zwischen Sätzen oder nach Ordnungszahlen", rule.GetDescription())
	require.Contains(t, rule.GetURL(), "grammatik-leerzeichen")
}
