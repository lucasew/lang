package ru

// Twin of languagetool-language-modules/ru/.../MorfologikRussianSpellerRuleTest.java
// Full ru_RU.dict deferred — map inject.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikRussianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikRussianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikRussianSpellerRule()
	require.Equal(t, MorfologikRussianSpellerRuleID, r.GetID())
	require.Equal(t, RussianSpellerDict, r.GetFileName())

	sp := morfologik.NewMorfologikSpeller(RussianSpellerDict, 1)
	for _, w := range []string{"привет", "мир", "тест"} {
		sp.AddWord(w)
	}
	sp.Suggestions["привт"] = []string{"привет"}
	// Map-inject unit path: clear initSpeller Multis so Speller map is used.
	r.ClearMultiSpellers()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled

	m, err := r.Match(languagetool.AnalyzePlain("привет мир"))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("привт"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "привет")

	// digits / punctuation not flagged
	m, err = r.Match(languagetool.AnalyzePlain("123"))
	require.NoError(t, err)
	require.Empty(t, m)
}
