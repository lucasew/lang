package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianWordCoherencyRuleTest.java
// Production: file pairs only (no invent case suffixes). Inflected forms via lemmas.
import (
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// twinCoherencyLemmas: surface → lemma (Java morph; ноль/нуль pair in coherency.txt).
var twinCoherencyLemmas = map[string]string{
	"ноль": "ноль", "нуль": "нуль",
	"нулю": "нуль", "нолю": "ноль",
	"нуля": "нуль", "ноля": "ноль",
}

func analyzeRU(s string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(s, ruCoherencyTagWord)
}

func analyzeRUText(s string) []*languagetool.AnalyzedSentence {
	// Multi-sentence like AnalyzeTextLocal with lemmas.
	if s == "" {
		return nil
	}
	// Paragraph split then sentence-local analyze
	paras := strings.Split(s, "\n\n")
	var out []*languagetool.AnalyzedSentence
	for _, para := range paras {
		if para == "" {
			continue
		}
		// single sentence chunks for these twin tests (or whole para)
		out = append(out, analyzeRU(para))
	}
	return out
}

func ruCoherencyTagWord(tok string) []languagetool.TokenTag {
	key := strings.ToLower(tok)
	key = strings.TrimFunc(key, func(r rune) bool { return !unicode.IsLetter(r) })
	if lem, ok := twinCoherencyLemmas[key]; ok {
		return []languagetool.TokenTag{{Lemma: lem}}
	}
	return nil
}

func TestRussianWordCoherencyRule_Rule(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		rule := NewRussianWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{analyzeRU(s)})), "good %q", s)
	}
	assertError := func(s string) {
		t.Helper()
		rule := NewRussianWordCoherencyRule(nil)
		require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{analyzeRU(s)})), "error %q", s)
	}
	assertGood("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C.")
	assertGood("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C.")
	assertError("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или ноль по шкале Кельвина.")
}

func TestRussianWordCoherencyRule_CallIndependence(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(NewRussianWordCoherencyRule(nil).MatchList([]*languagetool.AnalyzedSentence{analyzeRU(s)})))
	}
	assertGood("Абсолютный нуль.")
	assertGood("Ноль по шкале Кельвина.")
}

func TestRussianWordCoherencyRule_RuleCompleteTexts(t *testing.T) {
	check := func(s string) int {
		return len(NewRussianWordCoherencyRule(nil).MatchList(analyzeRUText(s)))
	}
	require.Equal(t, 0, check("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или нуль по шкале Кельвина."))
	require.Equal(t, 1, check("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или ноль по шкале Кельвина."))
	require.Equal(t, 1, check("Абсолютный нуль.\n\nСовсем недостижим. И ноль по шкале Кельвина."))
}

func TestRussianWordCoherencyRule_NoInventWithoutLemma(t *testing.T) {
	// Untagged: нулю not in file — must not invent soft-sign case expand.
	rule := NewRussianWordCoherencyRule(nil)
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или ноль по шкале Кельвина."),
	})))
}
