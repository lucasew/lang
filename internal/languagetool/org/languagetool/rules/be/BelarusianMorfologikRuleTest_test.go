package be

// Twin of BelarusianMorfologikRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of BelarusianMorfologikRuleTest.testMorfologikSpeller
func TestBelarusianMorfologikRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikBelarusianSpellerRule()
	require.Equal(t, MorfologikBelarusianSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikBelarusianSpellerRuleDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(MorfologikBelarusianSpellerRuleDict, 1)
	for _, w := range []string{"прывітанне", "свет", "мова"} {
		sp.AddWord(w)
	}
	sp.Suggestions["прывiтанне"] = []string{"прывітанне"} // latin i soft
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("прывітанне свет"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("xyzzy"))
	require.NoError(t, err)
	require.NotEmpty(t, m)
}
