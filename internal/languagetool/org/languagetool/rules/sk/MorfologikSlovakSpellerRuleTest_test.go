package sk

// Twin surface for MorfologikSlovakSpellerRule (Java has no dedicated *RuleTest class).
// Official resource: /sk/hunspell/sk_SK.dict (+ spelling.txt via InitSpellersFromGetters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.
// Analysis: AnalyzePlain (Slovak uses default word tokenizer; pure super has no letter-gate).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireSlovakSpellerDict skips when official sk_SK.dict is not discoverable (no invent).
func requireSlovakSpellerDict(t *testing.T) *MorfologikSlovakSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(SlovakSpellerDict) == "" {
		t.Skip("sk_SK.dict not discoverable (inspiration/.../sk/hunspell)")
	}
	r := NewMorfologikSlovakSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// Real-dict Match matrix: known good SK → 0; deliberate misspelling → ≥1.
func TestMorfologikSlovakSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireSlovakSpellerDict(t)
	require.Equal(t, MorfologikSlovakSpellerRuleID, r.GetID())
	require.Equal(t, SlovakSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_SK_SK", r.GetID())
	require.Equal(t, "/sk/hunspell/sk_SK.dict", r.GetFileName())

	matchN := func(s string) int {
		t.Helper()
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		return len(ms)
	}

	// correct words (official sk_SK.dict)
	require.Equal(t, 0, matchN("slovenský"))
	require.Equal(t, 0, matchN("ahoj"))
	require.Equal(t, 0, matchN("dom"))
	require.Equal(t, 0, matchN("človek"))
	require.Equal(t, 0, matchN("ďakujem"))
	// correct hyphenated form (also used by CompoundRule tests)
	require.Equal(t, 0, matchN("česko-slovenský"))

	// deliberate misspellings
	require.GreaterOrEqual(t, matchN("xxxxyzzyq"), 1)
	require.GreaterOrEqual(t, matchN("slovenskyy"), 1)
	require.GreaterOrEqual(t, matchN("domm"), 1)

	// digits / punctuation not flagged (no letters → skipped by parent Match)
	require.Equal(t, 0, matchN("123"))
	require.Equal(t, 0, matchN(","))
}
