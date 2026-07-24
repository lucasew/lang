package ga

// Twin of MorfologikIrishSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikIrishSpellerRuleTest.testMorfologikSpeller
func TestMorfologikIrishSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikIrishSpellerRule()
	require.Equal(t, MorfologikIrishSpellerRuleID, r.GetID())
	require.Equal(t, IrishSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(IrishSpellerDict, 1)
	for _, w := range []string{"dia", "duit", "Gaeilge"} {
		sp.AddWord(w)
	}
	sp.Suggestions["gaeilg"] = []string{"Gaeilge"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("dia duit"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("gaeilg"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "Gaeilge")
}
