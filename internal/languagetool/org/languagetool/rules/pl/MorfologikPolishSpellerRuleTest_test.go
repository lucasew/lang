package pl

// Twin of MorfologikPolishSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikPolishSpellerRuleTest.testMorfologikSpeller
func TestMorfologikPolishSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikPolishSpellerRule()
	require.Equal(t, MorfologikPolishSpellerRuleID, r.GetID())
	require.Equal(t, PolishSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(PolishSpellerDict, 1)
	for _, w := range []string{"cześć", "świat", "test"} {
		sp.AddWord(w)
	}
	sp.Suggestions["czesc"] = []string{"cześć"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("cześć świat"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("czesc"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "cześć")
}
