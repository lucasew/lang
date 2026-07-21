package ru

// Twin of MorfologikRussianYOSpellerRuleTest — ё-aware dict path inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikRussianYOSpellerRuleTest.testMorfologikSpeller
func TestMorfologikRussianYOSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikRussianYOSpellerRule()
	require.Equal(t, MorfologikRussianYOSpellerRuleID, r.GetID())
	require.Equal(t, RussianYOSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(RussianYOSpellerDict, 1)
	sp.AddWord("ёлка")
	sp.AddWord("елка") // YO dict often has both forms
	// Map-inject unit path: clear initSpeller Multis so Speller map is used.
	r.ClearMultiSpellers()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("ёлка"))
	require.NoError(t, err)
	require.Empty(t, m)

	// Latin tokens ignored by default (conf_ru_Value=0); use Cyrillic misspell
	m, err = r.Match(languagetool.AnalyzePlain("ёлкаа"))
	require.NoError(t, err)
	require.NotEmpty(t, m)
}
