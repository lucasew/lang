package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikBritishSpellerRule_Suggestions(t *testing.T) {
	r := NewMorfologikBritishSpellerRule()
	require.Equal(t, MorfologikBritishSpellerRuleID, r.GetID())
	sp := morfologik.NewMorfologikSpeller(BritishSpellerDict, 1)
	sp.AddWord("colour")
	sp.Suggestions["color"] = []string{"colour"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	require.Equal(t, []string{"colour"}, sp.FindReplacements("color"))
}

func TestMorfologikBritishSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikBritishSpellerRule()
	sp := morfologik.NewMorfologikSpeller(BritishSpellerDict, 1)
	sp.AddWord("hello")
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	ms, err := r.Match(languagetool.AnalyzePlain("hello helo"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
}
