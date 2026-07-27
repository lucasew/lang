package ml

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Metadata twin: getId / getFileName / Latin default (Java pure super, no isLatinScript override).
func TestMorfologikMalayalamSpellerRule(t *testing.T) {
	r := NewMorfologikMalayalamSpellerRule()
	require.Equal(t, MorfologikMalayalamSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikMalayalamSpellerRuleDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_ML_IN", r.GetID())
	require.Equal(t, "/ml/hunspell/ml_IN.dict", r.GetFileName())
	if r.SpellingCheckRule != nil {
		require.False(t, r.NonLatinScript)
	}
	if morfologik.DiscoverLanguageDict(MorfologikMalayalamSpellerRuleDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
}
