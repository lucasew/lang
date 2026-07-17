package nl

// Twin of MorfologikDutchSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikDutchSpellerRuleTest.testSpeller
func TestMorfologikDutchSpellerRule_Speller(t *testing.T) {
	r := NewMorfologikDutchSpellerRule()
	require.Equal(t, MorfologikDutchSpellerRuleID, r.GetID())
	require.Equal(t, DutchSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(DutchSpellerDict, 1)
	for _, w := range []string{"hallo", "wereld", "test"} {
		sp.AddWord(w)
	}
	sp.Suggestions["helo"] = []string{"hallo"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("hallo wereld"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("helo"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "hallo")
}
