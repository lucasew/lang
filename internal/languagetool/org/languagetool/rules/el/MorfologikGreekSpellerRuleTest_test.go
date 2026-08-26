package el

// Twin surface for MorfologikGreekSpellerRule (Java has no dedicated *RuleTest class).
// Official resource: /el/hunspell/el_GR.dict (+ spelling.txt via InitSpellersFromGetters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.
// Analysis: AnalyzePlain (Greek pure super; isLatinScript false → NonLatinScript).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireGreekSpellerDict skips when official el_GR.dict is not discoverable (no invent).
func requireGreekSpellerDict(t *testing.T) *MorfologikGreekSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(GreekSpellerDict) == "" {
		t.Skip("el_GR.dict not discoverable (inspiration/.../el/hunspell)")
	}
	r := NewMorfologikGreekSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// Real-dict Match matrix: known good EL → 0; deliberate misspelling → ≥1.
func TestMorfologikGreekSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireGreekSpellerDict(t)
	require.Equal(t, MorfologikGreekSpellerRuleID, r.GetID())
	require.Equal(t, GreekSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_EL_GR", r.GetID())
	require.Equal(t, "/el/hunspell/el_GR.dict", r.GetFileName())
	require.False(t, r.IsLatinScript())
	require.True(t, r.NonLatinScript)

	matchN := func(s string) int {
		t.Helper()
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		return len(ms)
	}

	// correct words (official el_GR.dict)
	require.Equal(t, 0, matchN("και"))
	require.Equal(t, 0, matchN("είναι"))
	require.Equal(t, 0, matchN("κείμενο"))
	require.Equal(t, 0, matchN("Ελλάδα"))
	require.Equal(t, 0, matchN("γλώσσα"))
	require.Equal(t, 0, matchN("άνθρωπος"))
	// spelling.txt extension also accepted via InitSpellersFromGetters
	require.Equal(t, 0, matchN("ιστότοπος"))

	// deliberate misspellings
	require.GreaterOrEqual(t, matchN("xxxxyzzyq"), 1)
	require.GreaterOrEqual(t, matchN("καιι"), 1)
	require.GreaterOrEqual(t, matchN("ελλααδα"), 1)

	// digits / punctuation not flagged (no letters → skipped by parent Match)
	require.Equal(t, 0, matchN("123"))
	require.Equal(t, 0, matchN(","))
}
