package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanFillerWordsRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanFillerWordsRule_Rule(t *testing.T) {
	// Java default minPercent=8 via UserConfig null
	rule := NewGermanFillerWordsRule(nil)
	require.Equal(t, "FILLER_WORDS_DE", rule.GetID())
	require.Equal(t, 8, rule.MinPercent)
	require.True(t, rule.ExcludeDirectSpeech)
	require.False(t, rule.TestTwoFollowing)
	require.False(t, rule.TestManyInSentence)
	require.Equal(t, "Statistische Stilanalyse: Füllwörter", rule.GetDescription())
	require.True(t, rule.IsDefaultOff())

	// more than 8% filler words (default)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Der Satz enthält augenscheinlich ein Füllwort."))))
	require.Equal(t, 2, len(rule.Match(languagetool.AnalyzePlain("Der Satz enthält augenscheinlich relativ viele Füllwörter."))))
	// less than 8% — don't show
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(
		"Der Satz enthält augenscheinlich ein Füllwort, aber es sind nicht genug um angezeigt zu werden."))))
	// direct speech / citation — don't show
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("»Der Satz enthält augenscheinlich ein Füllwort«"))))

	// multi-sentence under 8% with three-in-sentence / two-following OFF → 0
	longUnder := "Der Text enthält zu wenige Füllwörter, daher werden sie nicht angezeigt. " +
		"Was sich an diesem Satz mit diesem relativ einfachen Füllwort zeigt. " +
		"Dazu müssen noch eine Reihe von Sätzen geschrieben werden, um die Anzahl der Wörter zu erhöhen. " +
		"Langsam sollten die Anzahl der Worte für das Drücken unter die kritische Grenze reichen. " +
		"Jetzt schreibe ich allerdings einen Satz, der drei Füllwörter enthält, was allemal ziemlich ausreichend ist."
	require.Equal(t, 0, len(rule.MatchList(languagetool.AnalyzeTextLocal(longUnder))))

	longTwo := "Der Text enthält zu wenige Füllwörter, daher werden sie nicht angezeigt. " +
		"Was sich an diesem Satz mit diesem relativ einfachen Füllwort zeigt. " +
		"Dazu müssen noch eine Reihe von Sätzen geschrieben werden, um die Anzahl der Wörter zu erhöhen. " +
		"Langsam sollten die Anzahl der Worte für das Drücken unter die kritische Grenze reichen. " +
		"Jetzt schreibe ich einen Satz, der zwei Füllwörter hintereinander enthält, was allemal ziemlich ausreichend ist."
	require.Equal(t, 0, len(rule.MatchList(languagetool.AnalyzeTextLocal(longTwo))))

	// percentage set to zero — show all fillers (incl. speech: minPercent==0 path)
	rule0 := NewGermanFillerWordsRuleWithMinPercent(nil, 0)
	require.Equal(t, 1, len(rule0.Match(languagetool.AnalyzePlain("»Der Satz enthält augenscheinlich ein Füllwort«"))))
	require.Equal(t, 1, len(rule0.Match(languagetool.AnalyzePlain(
		"Der Satz enthält augenscheinlich ein Füllwort, aber es sind nicht genug um angezeigt zu werden."))))

	// sentence-start exception (num==1)
	require.Equal(t, 0, len(rule0.Match(languagetool.AnalyzePlain("Allerdings war es kalt."))))
	require.Equal(t, 1, len(rule0.Match(languagetool.AnalyzePlain("Es war allerdings kalt."))))
	// two-word exception: immer wieder
	require.Equal(t, 0, len(rule0.Match(languagetool.AnalyzePlain("Das passiert immer wieder hier."))))
	// after comma exception
	require.Equal(t, 0, len(rule0.Match(languagetool.AnalyzePlain("Es war kalt, allerdings regnete es."))))
}

func TestGermanFillerWordsRule_SentenceOptions(t *testing.T) {
	// Java UserConfig {8, true, true, true}: two-following + many-in-sentence
	// (Java getManyInSentence reads cf[2] same as two-following — both true.)

	// Isolated last-sentence: three fillers (allerdings, allemal, ziemlich).
	// Java many-in-sentence skips n±1 when counting others, so adjacent allemal/ziemlich
	// need two-following for sentence messages; both options on (Java cf[2] for both).
	last := "Jetzt schreibe ich allerdings einen Satz, der vier Füllwörter enthält, was allemal ziemlich ausreichend ist."
	ruleBothLast := NewGermanFillerWordsRule(nil)
	ruleBothLast.TestManyInSentence = true
	ruleBothLast.TestTwoFollowing = true
	ms := ruleBothLast.Match(languagetool.AnalyzePlain(last))
	require.Equal(t, 3, len(ms), "three fillers with both sentence options")
	for _, m := range ms {
		require.True(t,
			strings.Contains(m.GetMessage(), "Mehr als zwei potentielle Füllwörter") ||
				strings.Contains(m.GetMessage(), "Zwei potentielle Füllwörter hintereinander"),
			"msg=%q", m.GetMessage())
	}

	// many-in-sentence alone: non-adjacent trio → all get "mehr als zwei" message
	// (each sees 2 others beyond n±1)
	trio := "Es war allerdings heute allemal und dann ziemlich kalt draußen gewesen."
	// fillers: allerdings, allemal, ziemlich — separated by ≥1 non-filler
	ruleMany := NewGermanFillerWordsRule(nil)
	ruleMany.TestManyInSentence = true
	ms = ruleMany.Match(languagetool.AnalyzePlain(trio))
	require.Equal(t, 3, len(ms), "non-adjacent trio → many-in-sentence")
	for _, m := range ms {
		require.Contains(t, m.GetMessage(), "Mehr als zwei potentielle Füllwörter")
	}

	// two consecutive: allemal + ziemlich
	ruleTwo := NewGermanFillerWordsRule(nil)
	ruleTwo.TestTwoFollowing = true
	pair := "Das war allemal ziemlich ausreichend."
	ms = ruleTwo.Match(languagetool.AnalyzePlain(pair))
	require.Equal(t, 2, len(ms), "two consecutive fillers")
	for _, m := range ms {
		require.Contains(t, m.GetMessage(), "Zwei potentielle Füllwörter hintereinander")
	}

	// Multi-sentence under 8% with BOTH options: sentence-condition hits always emit.
	// Hits: relativ (deferred, under threshold) + allerdings/allemal/ziemlich (immediate).
	// Immediate count = 3 (Java full-text expects 4 when percent also emits relativ —
	// with AnalyzeTextLocal wordCount, 4/67≈6% < 8 so relativ stays deferred).
	four := "Der Text enthält zu wenige Füllwörter, daher werden sie nicht angezeigt. " +
		"Was sich an diesem Satz mit diesem relativ einfachen Füllwort zeigt. " +
		"Dazu müssen noch eine Reihe von Sätzen geschrieben werden, um die Anzahl der Wörter zu erhöhen. " +
		"Langsam sollten die Anzahl der Worte für das Drücken unter die kritische Grenze reichen. " +
		"Jetzt schreibe ich allerdings einen Satz, der vier Füllwörter enthält, was allemal ziemlich ausreichend ist."
	ruleBoth := NewGermanFillerWordsRule(nil)
	ruleBoth.TestTwoFollowing = true
	ruleBoth.TestManyInSentence = true
	ms = ruleBoth.MatchList(languagetool.AnalyzeTextLocal(four))
	require.Equal(t, 3, len(ms), "sentence-option immediate hits (deferred relativ under 8%)")
	// All messages are sentence-option style, not the percent limit message
	for _, m := range ms {
		require.True(t,
			strings.Contains(m.GetMessage(), "Füllwörter hintereinander") ||
				strings.Contains(m.GetMessage(), "Mehr als zwei potentielle Füllwörter"),
			"msg=%q", m.GetMessage())
	}

	// When minPercent is low enough that deferred also emit → 4 total (relativ + 3)
	ruleLow := NewGermanFillerWordsRuleWithMinPercent(nil, 0)
	ruleLow.TestTwoFollowing = true
	ruleLow.TestManyInSentence = true
	// MinPercent==0 still counts speech-path; all condition hits emit via percent>0 or sentence
	// With min 0: percent > 0 if any hits → deferred also shown. Sentence hits immediate too.
	// Avoid double-count: sentence hits go immediate; deferred only non-sentence.
	// relativ is deferred; 3 are immediate → MatchList size 4
	ms = ruleLow.MatchList(languagetool.AnalyzeTextLocal(four))
	require.Equal(t, 4, len(ms), "minPercent=0 emits deferred relativ + 3 sentence hits")
}

func TestGermanFillerWordsRule_LimitMessage(t *testing.T) {
	r := NewGermanFillerWordsRule(nil)
	require.Contains(t, r.getLimitMessage(0, 0), "könnte ein Füllwort")
	require.Contains(t, r.getLimitMessage(8, 12.4), "Mehr als 8%")
	require.Contains(t, r.getLimitMessage(8, 12.4), "12%")
}
