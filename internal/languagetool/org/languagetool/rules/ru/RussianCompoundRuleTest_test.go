package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianCompoundRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestRussianCompoundRule_Metadata(t *testing.T) {
	r := NewRussianCompoundRule(nil)
	require.Equal(t, "RU_COMPOUNDS", r.GetID())
	require.Equal(t, "Правописание через дефис", r.GetDescription())
	// Java AbstractCompoundRule ctor: Categories.MISC + ITSIssueType.Misspelling
	require.Equal(t, rules.CatMisc.GetCategory(nil).GetID(), r.GetCategory().GetID())
	require.Equal(t, rules.ITSMisspelling, r.GetLocQualityIssueType())
	// Java 3-arg super → shortDesc null
	require.Equal(t, "", r.ShortDesc)
	require.True(t, r.SentenceStartsWithUpperCase)
	// Java example pair
	require.Equal(t, []string{"конференц-зале"}, r.GetIncorrectExamples()[0].GetCorrections())
	require.Contains(t, r.GetIncorrectExamples()[0].GetExample(), "конференц зале")
}

func TestRussianCompoundRule_Rule(t *testing.T) {
	rule := NewRussianCompoundRule(nil)
	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(), "text %q", text)
		}
	}

	// correct sentences:
	check(0, "Он вышел из-за дома.")
	check(0, "Разработка ПО за идею.")
	// Both suggestion for some words:
	check(0, "естественно-научный")

	// incorrect sentences:
	check(1, "из за", "из-за")
	check(1, "по за", "по-за")
	check(1, "нет нет из за да да")

	// FIXME in Java: suggestions / longest match
	check(1, "Ростов на Дону", "Ростов-на-Дону")
	check(1, "Ростов на Дону — крупнейший город на юге Российской Федерации, административный центр Южного федерального округа и Ростовской области.")

	// no hyphen suggestion for some words:
	check(1, "кругло суточный", "круглосуточный")

	// also must not accept incorrect upper/lowercase spelling:
	check(0, "Ростов на дону")
	check(0, "Ведь сейчас в лос Анджелесе")

	// also detect an error if only some of the hyphens are missing:
	check(1, "Ростов-на Дону", "Ростов-на-Дону")

	// first part is a single character:
	check(0, "во-первых")
	check(1, "во первых", "во-первых")
	check(1, "Лос Анджелес", "Лос-Анджелес")
	check(1, "Ведь сейчас в Лос Анджелесе")
	check(1, "Ведь сейчас в Лос Анджелесе хорошая погода.")
	check(1, "Во первых, мы были довольно высоко над уровнем моря.")
	check(1, "Мы, во первых, были довольно высоко над уровнем моря.")
}
