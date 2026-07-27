package br

// Metadata surface for MorfologikBretonSpellerRule (Java twin outcomes live in
// MorfologikBretonSpellerRuleTest_test.go with real br_FR.dict — no map-inject invent).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikBretonSpellerRule_Metadata(t *testing.T) {
	r := NewMorfologikBretonSpellerRule()
	require.Equal(t, MorfologikBretonSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikBretonSpellerRuleDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_BR_FR", r.GetID())
	require.Equal(t, "/br/hunspell/br_FR.dict", r.GetFileName())
	// Java MorfologikBretonSpellerRule ctor: setIgnoreTaggedWords().
	require.True(t, r.IgnoreTaggedWords)
	// Java tokenizingPattern(): Pattern.compile("-")
	require.NotNil(t, r.TokenizingPattern)
	require.True(t, r.TokenizingPattern.MatchString("foo-bar"))
	require.False(t, r.TokenizingPattern.MatchString("foobar"))
	if morfologik.DiscoverLanguageDict(MorfologikBretonSpellerRuleDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
}
