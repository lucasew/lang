package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/LongSentenceRuleTest.java
// (extends core LongSentenceRuleTest helpers; DE overrides testMatch with maxWords=6).
//
// Java:
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("de-DE"));
//   LongSentenceRule rule = new LongSentenceRule(TestTools.getMessages("de"), new UserConfig(), 6);
//   assertNoMatch / assertMatch via rule.match(lt.analyzeText(input)) with exact fromPos/toPos.
import (
	"fmt"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// deMessagesLongSentence ports TestTools.getMessages("de") keys used by LongSentenceRule.
func deMessagesLongSentence() map[string]string {
	return map[string]string{
		// MessagesBundle_de: long_sentence_rule_msg2 (MessageFormat {0}).
		"long_sentence_rule_msg2": "Dieser Satz hat mehr als {0} Wörter. Kürzen Sie den Satz oder teilen Sie ihn, um die Lesbarkeit zu verbessern.",
		"long_sentence_rule_desc": "Lesbarkeit: Satz mit mehr als {0} Wörtern",
		"guiLongSentencesText":    "Maximum sentences length (5 - 100 words):",
	}
}

// matchDELongSentence ports rule.match(lt.analyzeText(text)).
func matchDELongSentence(rule *LongSentenceRule, lt *languagetool.JLanguageTool, text string) []*rules.RuleMatch {
	return rule.MatchList(lt.Analyze(text))
}

func TestLongSentenceRule_Match(t *testing.T) {
	// Java: new LongSentenceRule(TestTools.getMessages("de"), new UserConfig(), 6)
	rule := NewLongSentenceRule(deMessagesLongSentence(), languagetool.NewUserConfig(), 6)
	require.Equal(t, "TOO_LONG_SENTENCE_DE", rule.GetID())
	require.Equal(t, "Findet lange Sätze", rule.GetDescription())
	require.Equal(t, "Langer Satz", rule.GetShortMessage())
	require.Equal(t, "Langer Satz", rule.ShortMsg)
	require.NotEmpty(t, rule.GetIncorrectExamples())
	require.NotEmpty(t, rule.GetCorrectExamples())
	require.Equal(t, 6, rule.MaxWords)
	require.True(t, rule.HasTag(rules.TagPicky))
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.Equal(t, 0, rule.MinToCheckParagraph())

	// Java: new JLanguageTool(Languages.getLanguageForShortCode("de-DE"))
	lt := languagetool.NewJLanguageTool("de-DE")

	assertNoMatch := func(input string) {
		t.Helper()
		ms := matchDELongSentence(rule, lt, input)
		require.Equal(t, 0, len(ms), "assertNoMatch %q got %v", input, lsrSpans(ms))
	}
	assertMatch := func(input string, from, to int) {
		t.Helper()
		ms := matchDELongSentence(rule, lt, input)
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

// UserConfig.getConfigValueByID(getId()) with DE id TOO_LONG_SENTENCE_DE (Java polymorphic getId in super).
func TestLongSentenceRule_UserConfigMaxWords(t *testing.T) {
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID("TOO_LONG_SENTENCE_DE", []any{10})
	rule := NewLongSentenceRule(deMessagesLongSentence(), uc, 40)
	require.Equal(t, 10, rule.MaxWords)
	// Core id must not override DE id (parent would use TOO_LONG_SENTENCE if miswired).
	uc2 := languagetool.NewUserConfig()
	uc2.SetConfigValueByID(rules.LongSentenceRuleID, []any{7})
	rule2 := NewLongSentenceRule(deMessagesLongSentence(), uc2, 40)
	require.Equal(t, 40, rule2.MaxWords)
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
