package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanUnpairedQuotesRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanUnpairedQuotesRule_GermanRule(t *testing.T) {
	rule := NewGermanUnpairedQuotesRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	// correct
	require.Equal(t, 0, matchN("»Das sind die Sätze, die sie testen sollen«."))
	require.Equal(t, 0, matchN("«Das sind die ‹Sätze›, die sie testen sollen»."))
	require.Equal(t, 0, matchN("»Das sind die ›Sätze‹, die sie testen sollen«."))
	require.Equal(t, 0, matchN("„Das sind die Sätze, die sie testen sollen.“ „Hier steht ein zweiter Satz.“"))
	// incorrect
	require.Equal(t, 1, matchN("Die „Sätze zum Testen."))
	require.Equal(t, 1, matchN("Die «Sätze zum Testen."))
	require.Equal(t, 1, matchN("Die »Sätze zum Testen."))
}
