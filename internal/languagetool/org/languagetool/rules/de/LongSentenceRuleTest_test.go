package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/LongSentenceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestLongSentenceRule_Match(t *testing.T) {
	rule := NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "Dieser Satz hat mehr als %d Wörter",
	}, 6)
	require.Equal(t, "TOO_LONG_SENTENCE_DE", rule.GetID())
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Eins zwei drei vier fünf sechs."),
	})))
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Eins zwei drei vier fünf sechs sieben."),
	})
	require.Equal(t, 1, len(matches))
}
