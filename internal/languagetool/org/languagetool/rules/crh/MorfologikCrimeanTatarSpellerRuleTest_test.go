package crh

// Twin of MorfologikCrimeanTatarSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikCrimeanTatarSpellerRuleTest.testMorfologikSpeller
func TestMorfologikCrimeanTatarSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikCrimeanTatarSpellerRule()
	require.Equal(t, MorfologikCrimeanTatarSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikCrimeanTatarSpellerRuleDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(MorfologikCrimeanTatarSpellerRuleDict, 1)
	for _, w := range []string{"selâm", "qırım", "tili"} {
		sp.AddWord(w)
	}
	sp.Suggestions["selam"] = []string{"selâm"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("selâm qırım"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("selam"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "selâm")
}
