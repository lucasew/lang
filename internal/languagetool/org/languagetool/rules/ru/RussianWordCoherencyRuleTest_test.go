package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianWordCoherencyRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussianWordCoherencyRule_Rule(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		rule := NewRussianWordCoherencyRule(nil)
		require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)})), "good %q", s)
	}
	assertError := func(s string) {
		t.Helper()
		rule := NewRussianWordCoherencyRule(nil)
		require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)})), "error %q", s)
	}
	assertGood("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C.")
	assertGood("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C.")
	assertError("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или ноль по шкале Кельвина.")
}

func TestRussianWordCoherencyRule_CallIndependence(t *testing.T) {
	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(NewRussianWordCoherencyRule(nil).MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)})))
	}
	assertGood("Абсолютный нуль.")
	assertGood("Ноль по шкале Кельвина.")
}

func TestRussianWordCoherencyRule_RuleCompleteTexts(t *testing.T) {
	check := func(s string) int {
		return len(NewRussianWordCoherencyRule(nil).MatchList(languagetool.AnalyzeTextLocal(s)))
	}
	require.Equal(t, 0, check("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или нуль по шкале Кельвина."))
	require.Equal(t, 1, check("По шкале Цельсия абсолютному нулю соответствует температура −273,15 °C или ноль по шкале Кельвина."))
	require.Equal(t, 1, check("Абсолютный нуль.\n\nСовсем недостижим. И ноль по шкале Кельвина."))
}
