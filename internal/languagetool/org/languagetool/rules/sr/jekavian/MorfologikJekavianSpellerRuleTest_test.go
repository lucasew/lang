package jekavian

// Twin of org.languagetool.rules.sr.jekavian.MorfologikJekavianSpellerRuleTest.
// Official resource: /sr/dictionary/jekavian/serbian.dict (+ spelling/ignored/prohibit via getters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.
// Analysis: AnalyzePlain (default word tokenizer).
//
// Java MorfologikJekavianSpellerRule does not override isLatinScript (default true):
// pure non-Latin tokens are ignored by SpellingCheckRule.ignoreWord (pHasNoLetterLatin).
// Java test sentences are still 0-match outcomes (ignored or accepted); do not invent
// NonLatinScript=true for Serbian.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireJekavianSpellerDict skips when official jekavian serbian.dict is not discoverable (no invent).
func requireJekavianSpellerDict(t *testing.T) *MorfologikJekavianSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(JekavianSpellerDict) == "" {
		t.Skip("serbian.dict not discoverable (inspiration/.../sr/dictionary/jekavian)")
	}
	r := NewMorfologikJekavianSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// Port of MorfologikJekavianSpellerRuleTest.testMorfologikSpeller + metadata.
func TestMorfologikJekavianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireJekavianSpellerDict(t)
	require.Equal(t, MorfologikJekavianSpellerRuleID, r.GetID())
	require.Equal(t, JekavianSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_SR_JEKAVIAN", r.GetID())
	require.Equal(t, "/sr/dictionary/jekavian/serbian.dict", r.GetFileName())
	require.Equal(t, "/sr/dictionary/jekavian/ignored.txt", r.GetIgnoreFileName())
	require.Equal(t, "/sr/dictionary/jekavian/spelling.txt", r.GetSpellingFileName())
	require.Equal(t, "/sr/dictionary/jekavian/prohibit.txt", r.GetProhibitFileName())

	matchN := func(s string) int {
		t.Helper()
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		return len(ms)
	}

	// Java MorfologikJekavianSpellerRuleTest.testMorfologikSpeller — correct sentences
	require.Equal(t, 0, matchN("Тамо је лијеп цвијет."))
	require.Equal(t, 0, matchN("Дјечак и дјевојчица играју се заједно."))
	// Punctuation
	require.Equal(t, 0, matchN(","))
	// Roman numerals
	require.Equal(t, 0, matchN("III"))
}

// Port of MorfologikJekavianSpellerRuleTest.testSpellingCheck
func TestMorfologikJekavianSpellerRule_SpellingCheck(t *testing.T) {
	r := requireJekavianSpellerDict(t)
	ms, err := r.Match(languagetool.AnalyzePlain("Misspelled."))
	require.NoError(t, err)
	// Java: assertEquals(1, rule.match(...).length)
	require.Equal(t, 1, len(ms))
}
