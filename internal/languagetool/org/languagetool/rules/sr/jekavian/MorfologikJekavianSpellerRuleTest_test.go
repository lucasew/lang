package jekavian

// Twin of MorfologikJekavianSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikJekavianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikJekavianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikJekavianSpellerRule()
	require.Equal(t, MorfologikJekavianSpellerRuleID, r.GetID())
	require.Equal(t, JekavianSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(JekavianSpellerDict, 1)
	for _, w := range []string{"здраво", "свијет", "тест"} {
		sp.AddWord(w)
	}
	sp.Suggestions["свијетт"] = []string{"свијет"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("здраво свијет"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("свијетт"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "свијет")
}
