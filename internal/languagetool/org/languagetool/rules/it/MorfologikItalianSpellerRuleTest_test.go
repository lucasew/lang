package it

// Twin of MorfologikItalianSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikItalianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikItalianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikItalianSpellerRule()
	require.Equal(t, MorfologikItalianSpellerRuleID, r.GetID())
	require.Equal(t, ItalianSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(ItalianSpellerDict, 1)
	for _, w := range []string{"ciao", "mondo", "prova"} {
		sp.AddWord(w)
	}
	sp.Suggestions["ciao"] = nil
	sp.Suggestions["ciaoo"] = []string{"ciao"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("ciao mondo"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("ciaoo"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "ciao")
}
