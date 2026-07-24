package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/LongSentenceRuleTest.java
// (extends core LongSentenceRuleTest helpers; DE overrides testMatch with maxWords=6).
import (
	"fmt"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestLongSentenceRule_Match(t *testing.T) {
	// Java: new LongSentenceRule(messages, new UserConfig(), 6)
	rule := NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "Dieser Satz hat mehr als %d Wörter",
	}, 6)
	require.Equal(t, "TOO_LONG_SENTENCE_DE", rule.GetID())
	require.Equal(t, "Findet lange Sätze", rule.GetDescription())
	require.Equal(t, "Langer Satz", rule.ShortMsg)
	require.NotEmpty(t, rule.GetIncorrectExamples())

	assertNoMatch := func(input string) {
		t.Helper()
		ms := rule.MatchList(languagetool.AnalyzeTextLocal(input))
		require.Equal(t, 0, len(ms), "assertNoMatch %q got %v", input, lsrSpans(ms))
	}
	assertMatch := func(input string, from, to int) {
		t.Helper()
		ms := rule.MatchList(languagetool.AnalyzeTextLocal(input))
		require.Equal(t, 1, len(ms), "assertMatch %q got %d spans %v", input, len(ms), lsrSpans(ms))
		require.Equal(t, from, ms[0].GetFromPos(), "fromPos for %q", input)
		require.Equal(t, to, ms[0].GetToPos(), "toPos for %q", input)
	}

	// correct / short (Java assertNoMatch)
	assertNoMatch("Eins zwei drei vier fünf sechs.")
	// Words after colon are treated like a separate sentence (and quoted short-circuit)
	assertNoMatch("Ich zähle jetzt: \"Eins zwei drei vier fünf sechs.\"")
	// quotedSentEnd short-circuit ([?!.]["“”„»«])
	assertNoMatch("Peter, bist du bereit?” Er nickte nur.\n")
	assertNoMatch("Peter, du bist bereit.” Er nickte nur.\n")
	// parentheses exclude words inside quotes/parens (indexOfQuote)
	assertNoMatch("Eins zwei drei vier fünf (sechs sieben) acht.")

	// incorrect (Java assertMatch with exact spans)
	assertMatch("Eins zwei drei vier fünf sechs sieben.", 0, 38)
	assertMatch("Eins zwei drei vier fünf (sechs sieben) acht neun.", 0, 50)
	// Java: fromPosToken is NOT reset after ':' → span starts at first word "Ich"
	assertMatch("Ich zähle jetzt: Eins zwei drei vier fünf sechs sieben.", 0, 55)
	// multi-sentence: pos offset of second sentence (Java lt.analyzeText)
	assertMatch("Ein Satz. Eins zwei drei vier fünf sechs sieben.", 10, 48)
}

func lsrSpans(ms []*rules.RuleMatch) string {
	if len(ms) == 0 {
		return "[]"
	}
	s := "["
	for i, m := range ms {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%d-%d", m.GetFromPos(), m.GetToPos())
	}
	return s + "]"
}
