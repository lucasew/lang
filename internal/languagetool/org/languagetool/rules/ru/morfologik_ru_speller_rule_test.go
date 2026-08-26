package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikRussianSpellerRule(t *testing.T) {
	r := NewMorfologikRussianSpellerRule()
	// Java MorfologikRussianSpellerRule.RULE_ID / RESOURCE_FILENAME
	require.Equal(t, "MORFOLOGIK_RULE_RU_RU", MorfologikRussianSpellerRuleID)
	require.Equal(t, "/ru/hunspell/ru_RU.dict", RussianSpellerDict)
	require.Equal(t, MorfologikRussianSpellerRuleID, r.GetID())
	require.Equal(t, RussianSpellerDict, r.GetFileName())
	require.Equal(t, 0, r.ConfCheckLatin)
	// Java isLatinScript() = false → NonLatinScript
	require.True(t, r.NonLatinScript)
	// Java example pair
	require.Len(t, r.GetIncorrectExamples(), 1)
	require.Len(t, r.GetCorrectExamples(), 1)
	// Java getConfigureText / getRuleOptions
	require.Equal(t, "Проверять слова на латинице, только термины (0/1)", r.GetConfigureText())
	opts := r.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 0, opts[0].DefaultValue)
	require.Equal(t, r.GetConfigureText(), opts[0].ConfigureText)
	require.Equal(t, 0, opts[0].MinConfigurableValue)
	require.Equal(t, 1, opts[0].MaxConfigurableValue)
}

func TestMorfologikRussianSpellerRules(t *testing.T) {
	require.Equal(t, MorfologikRussianSpellerRuleID, NewMorfologikRussianSpellerRule().GetID())
	require.Equal(t, MorfologikRussianYOSpellerRuleID, NewMorfologikRussianYOSpellerRule().GetID())
}

func TestMorfologikRussianYOSpellerRule(t *testing.T) {
	r := NewMorfologikRussianYOSpellerRule()
	// Java MorfologikRussianYOSpellerRule.RULE_ID / RESOURCE_FILENAME
	require.Equal(t, "MORFOLOGIK_RULE_RU_RU_YO", MorfologikRussianYOSpellerRuleID)
	require.Equal(t, "/ru/hunspell/ru_RU_yo.dict", RussianYOSpellerDict)
	require.Equal(t, MorfologikRussianYOSpellerRuleID, r.GetID())
	require.Equal(t, RussianYOSpellerDict, r.GetFileName())
	require.Equal(t, "Проверка орфографии. Только «Ё» (экспериментальное правило).", r.GetDescription())
	// Java setDefaultOff()
	require.True(t, r.IsDefaultOff())
	require.Equal(t, 0, r.ConfCheckLatin)
	// Java isLatinScript() = false → NonLatinScript
	require.True(t, r.NonLatinScript)
	// Java example pair
	require.Len(t, r.GetIncorrectExamples(), 1)
	require.Len(t, r.GetCorrectExamples(), 1)
	// Java getConfigureText / getRuleOptions
	require.Equal(t, "Проверять слова на латинице, только термины (0/1)", r.GetConfigureText())
	opts := r.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 0, opts[0].DefaultValue)
	require.Equal(t, r.GetConfigureText(), opts[0].ConfigureText)
	require.Equal(t, 0, opts[0].MinConfigurableValue)
	require.Equal(t, 1, opts[0].MaxConfigurableValue)
}

func TestMorfologikRussianYOSpellerRule_UserConfigConfValue(t *testing.T) {
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID(MorfologikRussianYOSpellerRuleID, []any{1})
	r := NewMorfologikRussianYOSpellerRuleWithUser(uc)
	require.Equal(t, 1, r.ConfCheckLatin)
	hello := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("hello", nil, nil), 0)
	require.False(t, r.ruIgnoreToken([]*languagetool.AnalyzedTokenReadings{hello}, 0))
}

func TestRussianLettersGate(t *testing.T) {
	require.True(t, russianLetters.MatchString("привет"))
	require.True(t, russianLetters.MatchString("Ёлка"))
	require.True(t, russianLetters.MatchString("кто-то"))
	require.False(t, russianLetters.MatchString("hello"))
	require.False(t, russianLetters.MatchString("test123"))
	require.False(t, russianLetters.MatchString("mixкирилл"))
}

func TestRuIgnoreToken_LatinSkippedByDefault(t *testing.T) {
	r := NewMorfologikRussianSpellerRule()
	// Build token list like Java ignoreToken (not AnalyzePlain invent path for Match).
	hello := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("hello", nil, nil), 0)
	mir := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("мир", nil, nil), 0)
	toks := []*languagetool.AnalyzedTokenReadings{hello, mir}
	require.True(t, r.ruIgnoreToken(toks, 0), "Latin ignored when conf_ru_Value != 1")
	// Russian letters match → not letter-gate ignored; may still IgnoreWord
	require.True(t, russianLetters.MatchString("мир"))
	// IgnoreWordAt path for "мир" is false for normal dictionary word (no ignore set)
	require.False(t, r.ruIgnoreToken(toks, 1))
}

func TestRuIgnoreToken_LatinWhenConf1(t *testing.T) {
	r := NewMorfologikRussianSpellerRule()
	r.ConfCheckLatin = 1
	hello := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("hello", nil, nil), 0)
	toks := []*languagetool.AnalyzedTokenReadings{hello}
	// conf=1 → do not skip on letter gate; IgnoreWord("hello") may still false (has Latin letters
	// but NonLatinScript uses pHasNoLetter — hello has letters → not ignored via IgnoreWord)
	require.False(t, r.ruIgnoreToken(toks, 0))
}

func TestMorfologikRussianSpellerRule_UserConfigConfValue(t *testing.T) {
	uc := languagetool.NewUserConfig()
	uc.SetConfigValueByID(MorfologikRussianSpellerRuleID, []any{1})
	r := NewMorfologikRussianSpellerRuleWithUser(uc)
	require.Equal(t, 1, r.ConfCheckLatin)
	// conf=1: Latin not letter-gate ignored
	hello := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("hello", nil, nil), 0)
	require.False(t, r.ruIgnoreToken([]*languagetool.AnalyzedTokenReadings{hello}, 0))
}

func TestFilterNoSuggestWords(t *testing.T) {
	r := NewMorfologikRussianSpellerRule()
	got := r.filterNoSuggestWords([]string{"привет", "Блоггер", "дрочим", "мир"})
	require.Equal(t, []string{"привет", "мир"}, got)
}

func TestFilterNoSuggestWords_YO(t *testing.T) {
	r := NewMorfologikRussianYOSpellerRule()
	// Java YO NOSUGGEST has "елка" (е not ё); toLowerCase "Ёлка" → "ёлка" keeps it
	got := r.filterNoSuggestWords([]string{"елка", "Ёлка", "привет", "блоггер"})
	require.Equal(t, []string{"Ёлка", "привет"}, got)
}

func TestRussianNonLatinScript(t *testing.T) {
	require.True(t, NewMorfologikRussianSpellerRule().NonLatinScript)
	require.True(t, NewMorfologikRussianYOSpellerRule().NonLatinScript)
}

// Real-dict Latin gate: default conf skips Latin tokens entirely.
func TestMatch_LatinNotFlagged_RealDict(t *testing.T) {
	if morfologik.DiscoverLanguageDict(RussianSpellerDict) == "" {
		t.Skip("ru_RU.dict not discoverable")
	}
	r := NewMorfologikRussianSpellerRule()
	require.NotNil(t, r.Speller1)
	m, err := r.Match(analyzeRU("hello"))
	require.NoError(t, err)
	require.Empty(t, m)
}

// Real-dict NOSUGGEST surface: filterNoSuggestWords drops blocked suggestions when present.
func TestMatch_FiltersNoSuggest_Unit(t *testing.T) {
	r := NewMorfologikRussianSpellerRule()
	// Direct unit of filter path (Java filterNoSuggestWords) — no map-inject Match.
	got := r.filterNoSuggestWords([]string{"привет", "блоггер", "дрочим", "анальный", "орочем", "мир"})
	require.Equal(t, []string{"привет", "мир"}, got)
}
