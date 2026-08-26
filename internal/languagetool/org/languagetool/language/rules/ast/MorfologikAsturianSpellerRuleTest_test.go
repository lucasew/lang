package ast

// Twin of MorfologikAsturianSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikAsturianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikAsturianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikAsturianSpellerRule()
	require.Equal(t, MorfologikAsturianSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikAsturianSpellerRuleDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(MorfologikAsturianSpellerRuleDict, 1)
	for _, w := range []string{"hola", "asturies"} {
		sp.AddWord(w)
	}
	sp.Suggestions["asturie"] = []string{"asturies"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("hola asturies"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("asturie"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "asturies")
}
