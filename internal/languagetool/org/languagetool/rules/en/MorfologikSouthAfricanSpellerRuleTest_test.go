package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikSouthAfricanSpellerRule_Suggestions(t *testing.T) {
	r := NewMorfologikSouthAfricanSpellerRule()
	require.Equal(t, MorfologikSouthAfricanSpellerRuleID, r.GetID())
	sp := morfologik.NewMorfologikSpeller(SouthAfricanSpellerDict, 1)
	sp.AddWord("colour")
	sp.Suggestions["color"] = []string{"colour"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	require.Equal(t, []string{"colour"}, sp.FindReplacements("color"))
}

func TestMorfologikSouthAfricanSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikSouthAfricanSpellerRule()
	sp := morfologik.NewMorfologikSpeller(SouthAfricanSpellerDict, 1)
	sp.AddWord("favourite")
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("favourite favrite")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}
