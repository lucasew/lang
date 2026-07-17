package ekavian

// Twin of MorfologikEkavianSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikEkavianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikEkavianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikEkavianSpellerRule()
	require.Equal(t, MorfologikEkavianSpellerRuleID, r.GetID())
	require.Equal(t, EkavianSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(EkavianSpellerDict, 1)
	for _, w := range []string{"здраво", "свет", "тест"} {
		sp.AddWord(w)
	}
	sp.Suggestions["здрав"] = []string{"здраво"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("здраво свет"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("здрав"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "здраво")
}
