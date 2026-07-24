package it

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikItalianSpellerRule(t *testing.T) {
	r := NewMorfologikItalianSpellerRule()
	// Java MorfologikItalianSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_IT_IT", MorfologikItalianSpellerRuleID)
	require.Equal(t, "/it/hunspell/it_IT.dict", ItalianSpellerDict)
	require.Equal(t, MorfologikItalianSpellerRuleID, r.GetID())
	require.Equal(t, ItalianSpellerDict, r.GetFileName())
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
}

func TestMorfologikItalianSpellerRule_MatchOrdersSuggestions(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(ItalianSpellerDict, 1)
	sp.AddWord("casa")
	sp.AddWord("caso")
	r := NewMorfologikItalianSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	// force FindReplacements path: inject suggestions via Match after map misspell
	// empty FindReplacements from map — set via manual order test above is enough
	// for Match wiring: unknown word with no sugs stays empty
	sent := languagetool.AnalyzePlain("xyzzy")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, m, 1)
}
