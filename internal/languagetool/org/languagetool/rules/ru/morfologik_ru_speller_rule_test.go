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
	require.Contains(t, r.GetDescription(), "Ё")
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
	sent := languagetool.AnalyzePlain("hello мир")
	toks := sent.GetTokensWithoutWhitespace()
	for i, tok := range toks {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		if tok.GetToken() == "hello" {
			require.True(t, r.ruIgnoreToken(toks, i))
		}
		if tok.GetToken() == "мир" {
			// not ignored via letter gate (may still IgnoreWord)
			require.False(t, r.ruIgnoreToken(toks, i) && !r.IgnoreWord("мир"))
			// letter gate alone: russian matches → only IgnoreWord path
			require.True(t, russianLetters.MatchString("мир"))
		}
	}
}

func TestRuIgnoreToken_LatinWhenConf1(t *testing.T) {
	r := NewMorfologikRussianSpellerRule()
	r.ConfCheckLatin = 1
	sent := languagetool.AnalyzePlain("hello")
	toks := sent.GetTokensWithoutWhitespace()
	for i, tok := range toks {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		// conf=1 → do not skip on letter gate; IgnoreWord("hello") may still false
		require.False(t, r.ruIgnoreToken(toks, i))
	}
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

func TestMatch_LatinNotFlagged(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(RussianSpellerDict, 1)
	sp.AddWord("мир")
	r := NewMorfologikRussianSpellerRule()
	r.ClearMultiSpellers()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	m, err := r.Match(languagetool.AnalyzePlain("hello"))
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestMatch_FiltersNoSuggest(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(RussianSpellerDict, 1)
	sp.AddWord("привет")
	sp.Suggestions["привт"] = []string{"привет", "блоггер", "дрочим"}
	r := NewMorfologikRussianSpellerRule()
	r.ClearMultiSpellers()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	m, err := r.Match(languagetool.AnalyzePlain("привт"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Equal(t, []string{"привет"}, m[0].GetSuggestedReplacements())
}

func TestRussianNonLatinScript(t *testing.T) {
	require.True(t, NewMorfologikRussianSpellerRule().NonLatinScript)
	require.True(t, NewMorfologikRussianYOSpellerRule().NonLatinScript)
}
