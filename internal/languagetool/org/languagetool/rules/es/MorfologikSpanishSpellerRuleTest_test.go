package es

// Twin of MorfologikSpanishSpellerRuleTest — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikSpanishSpellerRuleTest.testMorfologikSpeller
func TestMorfologikSpanishSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikSpanishSpellerRule()
	require.Equal(t, MorfologikSpanishSpellerRuleID, r.GetID())
	require.Equal(t, SpanishSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(SpanishSpellerDict, 1)
	for _, w := range []string{"hola", "mundo", "prueba"} {
		sp.AddWord(w)
	}
	sp.Suggestions["ola"] = []string{"hola"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("hola mundo"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("ola"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "hola")

	// known word not misspelled
	require.False(t, r.IsMisspelled("hola"))
	require.True(t, r.IsMisspelled("xyzzy"))
}
