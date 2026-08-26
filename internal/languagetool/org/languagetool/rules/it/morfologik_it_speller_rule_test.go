package it

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikItalianSpellerRule_Metadata(t *testing.T) {
	r := NewMorfologikItalianSpellerRule()
	// Java MorfologikItalianSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_IT_IT", MorfologikItalianSpellerRuleID)
	require.Equal(t, "/it/hunspell/it_IT.dict", ItalianSpellerDict)
	require.Equal(t, MorfologikItalianSpellerRuleID, r.GetID())
	require.Equal(t, ItalianSpellerDict, r.GetFileName())
	// Latin script default (Java does not override isLatinScript).
	if r.SpellingCheckRule != nil {
		require.False(t, r.NonLatinScript)
	}
	if morfologik.DiscoverLanguageDict(ItalianSpellerDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
}

// Port of orderSuggestions: drop capitalized dup when lower also present.
func TestOrderItalianSuggestions_DropCapitalizedDup(t *testing.T) {
	// word not capitalized, list has both "casa" and "Casa" → drop "Casa"
	got := orderItalianSuggestions([]string{"casa", "Casa", "caso"}, "caza")
	require.Equal(t, []string{"casa", "caso"}, got)

	// word is capitalized → keep capitalized suggestion
	got = orderItalianSuggestions([]string{"casa", "Casa"}, "Caza")
	require.Equal(t, []string{"casa", "Casa"}, got)

	// capitalized sug without lowercase twin → keep
	got = orderItalianSuggestions([]string{"Roma", "romo"}, "romo")
	require.Equal(t, []string{"Roma", "romo"}, got)

	// real esmpio raw shape (Java findReplacements before orderSuggestions)
	got = orderItalianSuggestions(
		[]string{"Empio", "Esempio", "Espio", "Espiò", "empio", "esempio", "espio", "espiò"},
		"esmpio",
	)
	require.Equal(t, []string{"empio", "esempio", "espio", "espiò"}, got)
}
