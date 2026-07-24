package br

// Twin of MorfologikBretonSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikBretonSpellerRuleTest.testMorfologikSpeller
func TestMorfologikBretonSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikBretonSpellerRule()
	require.Equal(t, MorfologikBretonSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikBretonSpellerRuleDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(MorfologikBretonSpellerRuleDict, 1)
	for _, w := range []string{"demat", "brezhoneg", "test"} {
		sp.AddWord(w)
	}
	sp.Suggestions["dematt"] = []string{"demat"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("demat brezhoneg"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("dematt"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "demat")
}
