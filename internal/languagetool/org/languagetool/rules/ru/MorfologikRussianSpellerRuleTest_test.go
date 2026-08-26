package ru

// Twin of languagetool-language-modules/ru/.../MorfologikRussianSpellerRuleTest.java
// Official resource: /ru/hunspell/ru_RU.dict (+ spelling.txt via InitSpellersFromGetters).
// Analysis ports JLanguageTool(Russian).getAnalyzedSentence — RU tokenizer + tagger when present.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	tagru "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	rutok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ru"
	"github.com/stretchr/testify/require"
)

// requireRussianSpellerDict skips when official ru_RU.dict is not discoverable (no invent).
func requireRussianSpellerDict(t *testing.T) *MorfologikRussianSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(RussianSpellerDict) == "" {
		t.Skip("ru_RU.dict not discoverable (inspiration/.../ru/hunspell)")
	}
	r := NewMorfologikRussianSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// analyzeRU ports JLanguageTool(Russian.getInstance()).getAnalyzedSentence.
// Uses RussianWordTokenizer; wires RussianTagger when russian.dict is available.
func analyzeRU(text string) *languagetool.AnalyzedSentence {
	wt := rutok.NewRussianWordTokenizer()
	if tagru.DiscoverRussianPOSDict() != "" {
		tagru.EnsureDefaultRussianTagger()
		if tg := tagru.DefaultRussianTagger; tg != nil {
			lt := languagetool.NewJLanguageTool("ru")
			lt.SentenceTagger = tg.Tag
			if s := lt.GetAnalyzedSentence(text); s != nil {
				return s
			}
		}
	}
	return languagetool.AnalyzeWithTokenizer(text, wt)
}

// Port of MorfologikRussianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikRussianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireRussianSpellerDict(t)
	require.Equal(t, MorfologikRussianSpellerRuleID, r.GetID())
	require.Equal(t, RussianSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_RU_RU", r.GetID())
	require.Equal(t, "/ru/hunspell/ru_RU.dict", r.GetFileName())

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
	// correct word
	require.Equal(t, 0, matchN("ежик"))

	// incorrect word with hyphen
	require.Equal(t, 1, matchN("юго-зпдный"))
	// correct word with hyphen
	require.Equal(t, 0, matchN("северо-восточный"))

	// correct word with hyphen
	require.Equal(t, 0, matchN("Ростов-на-Дону"))
	// incorrect word with hyphen
	require.Equal(t, 1, matchN("Ростов-на-дону"))
}
