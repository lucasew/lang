package ml

// Twin of MorfologikMalayalamSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikMalayalamSpellerRuleTest.testMorfologikSpeller
func TestMorfologikMalayalamSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikMalayalamSpellerRule()
	require.Equal(t, MorfologikMalayalamSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikMalayalamSpellerRuleDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(MorfologikMalayalamSpellerRuleDict, 1)
	// use ASCII stand-ins for inject (full Malayalam dict deferred)
	for _, w := range []string{"namaskaram", "malayalam"} {
		sp.AddWord(w)
	}
	sp.Suggestions["namaskrm"] = []string{"namaskaram"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("namaskaram malayalam"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("namaskrm"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "namaskaram")
}
