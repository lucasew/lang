package tl

// Twin of MorfologikTagalogSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikTagalogSpellerRuleTest.testMorfologikSpeller
func TestMorfologikTagalogSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikTagalogSpellerRule()
	require.Equal(t, MorfologikTagalogSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikTagalogSpellerRuleDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(MorfologikTagalogSpellerRuleDict, 1)
	for _, w := range []string{"kumusta", "mundo"} {
		sp.AddWord(w)
	}
	sp.Suggestions["kumusta"] = nil
	sp.Suggestions["kumust"] = []string{"kumusta"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("kumusta mundo"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("kumust"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "kumusta")
}
