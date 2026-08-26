package ru

// Twin of languagetool-language-modules/ru/.../MorfologikRussianYOSpellerRuleTest.java
// Official resource: /ru/hunspell/ru_RU_yo.dict (+ spelling.txt via InitSpellersFromGetters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireRussianYOSpellerDict skips when official ru_RU_yo.dict is not discoverable (no invent).
func requireRussianYOSpellerDict(t *testing.T) *MorfologikRussianYOSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(RussianYOSpellerDict) == "" {
		t.Skip("ru_RU_yo.dict not discoverable (inspiration/.../ru/hunspell)")
	}
	r := NewMorfologikRussianYOSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// Port of MorfologikRussianYOSpellerRuleTest.testMorfologikSpeller
func TestMorfologikRussianYOSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireRussianYOSpellerDict(t)
	require.Equal(t, MorfologikRussianYOSpellerRuleID, r.GetID())
	require.Equal(t, RussianYOSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_RU_RU_YO", r.GetID())
	require.Equal(t, "/ru/hunspell/ru_RU_yo.dict", r.GetFileName())
	require.True(t, r.IsDefaultOff(), "Java setDefaultOff() in ctor")

	matchN := func(s string) int {
		t.Helper()
		ms, err := r.Match(analyzeRU(s))
		require.NoError(t, err)
		return len(ms)
	}

	// correct word
	require.Equal(t, 0, matchN("русский"))
	// correct word
	require.Equal(t, 0, matchN("ёжик"))
	// incorrect word (YO dict: e-form of hedgehog is wrong; differs from base ru_RU)
	require.Equal(t, 1, matchN("ежик"))

	// incorrect word with hyphen
	require.Equal(t, 1, matchN("юго-зпдный"))
	// correct word with hyphen
	require.Equal(t, 0, matchN("северо-восточный"))

	// correct word with hyphen
	require.Equal(t, 0, matchN("Ростов-на-Дону"))
	// Java comments out Ростов-на-дону (mistake in dict) — do not invent a hard fail.
}
