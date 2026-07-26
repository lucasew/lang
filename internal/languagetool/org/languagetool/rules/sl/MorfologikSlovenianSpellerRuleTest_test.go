package sl

// Twin surface for MorfologikSlovenianSpellerRule (Java has no dedicated *RuleTest class).
// Official resource: /sl/hunspell/sl_SI.dict (+ spelling.txt via InitSpellersFromGetters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.
// Analysis: AnalyzePlain (Slovenian uses default word tokenizer; pure super has no letter-gate).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireSlovenianSpellerDict skips when official sl_SI.dict is not discoverable (no invent).
func requireSlovenianSpellerDict(t *testing.T) *MorfologikSlovenianSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(SlovenianSpellerDict) == "" {
		t.Skip("sl_SI.dict not discoverable (inspiration/.../sl/hunspell)")
	}
	r := NewMorfologikSlovenianSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// Real-dict Match matrix: known good SL → 0; deliberate misspelling → ≥1.
func TestMorfologikSlovenianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireSlovenianSpellerDict(t)
	require.Equal(t, MorfologikSlovenianSpellerRuleID, r.GetID())
	require.Equal(t, SlovenianSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_SL_SI", r.GetID())
	require.Equal(t, "/sl/hunspell/sl_SI.dict", r.GetFileName())

	matchN := func(s string) int {
		t.Helper()
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		return len(ms)
	}

	// correct words (official sl_SI.dict)
	require.Equal(t, 0, matchN("hiša"))
	require.Equal(t, 0, matchN("dober"))
	require.Equal(t, 0, matchN("dan"))
	require.Equal(t, 0, matchN("človek"))
	require.Equal(t, 0, matchN("hvala"))
	require.Equal(t, 0, matchN("mesto"))
	// correct proper noun form (case-sensitive surface accepted by dict)
	require.Equal(t, 0, matchN("Slovenija"))

	// deliberate misspellings
	require.GreaterOrEqual(t, matchN("xxxxyzzyq"), 1)
	require.GreaterOrEqual(t, matchN("slovenijaa"), 1)
	require.GreaterOrEqual(t, matchN("domm"), 1)

	// digits / punctuation not flagged (no letters → skipped by parent Match)
	require.Equal(t, 0, matchN("123"))
	require.Equal(t, 0, matchN(","))
}
