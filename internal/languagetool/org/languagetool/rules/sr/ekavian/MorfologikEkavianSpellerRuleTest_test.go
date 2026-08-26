package ekavian

// Twin of org.languagetool.rules.sr.ekavian.MorfologikEkavianSpellerRuleTest.
// Official resource: /sr/dictionary/ekavian/serbian.dict (+ spelling/ignored/prohibit via getters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.
// Analysis: AnalyzePlain (default word tokenizer).
//
// Java MorfologikEkavianSpellerRule does not override isLatinScript (default true):
// pure non-Latin tokens are ignored by SpellingCheckRule.ignoreWord (pHasNoLetterLatin).
// Java test sentences are still 0-match outcomes (ignored or accepted); do not invent
// NonLatinScript=true for Serbian.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireEkavianSpellerDict skips when official ekavian serbian.dict is not discoverable (no invent).
func requireEkavianSpellerDict(t *testing.T) *MorfologikEkavianSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(EkavianSpellerDict) == "" {
		t.Skip("serbian.dict not discoverable (inspiration/.../sr/dictionary/ekavian)")
	}
	r := NewMorfologikEkavianSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// Port of MorfologikEkavianSpellerRuleTest.testMorfologikSpeller + metadata.
func TestMorfologikEkavianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireEkavianSpellerDict(t)
	require.Equal(t, MorfologikEkavianSpellerRuleID, r.GetID())
	require.Equal(t, EkavianSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_SR_EKAVIAN", r.GetID())
	require.Equal(t, "/sr/dictionary/ekavian/serbian.dict", r.GetFileName())
	require.Equal(t, "/sr/dictionary/ekavian/ignored.txt", r.GetIgnoreFileName())
	require.Equal(t, "/sr/dictionary/ekavian/spelling.txt", r.GetSpellingFileName())
	require.Equal(t, "/sr/dictionary/ekavian/prohibit.txt", r.GetProhibitFileName())

	matchN := func(s string) int {
		t.Helper()
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		return len(ms)
	}

	// Java MorfologikEkavianSpellerRuleTest.testMorfologikSpeller — correct sentences
	require.Equal(t, 0, matchN("Тамо је леп цвет"))
	require.Equal(t, 0, matchN("Дечак и девојчица играју се заједно."))
	// Punctuation
	require.Equal(t, 0, matchN(","))
	// Roman numerals
	require.Equal(t, 0, matchN("III"))

	// Deliberate misspelling (Latin surface): Java isLatinScript default true does not
	// ignore Latin letters; official multi treats nonsense as misspelled.
	// Do not assert Cyrillic бткие ≥1 — ignoreWord short-circuits pure non-Latin.
	require.GreaterOrEqual(t, matchN("xxxxyzzyq"), 1)
}
