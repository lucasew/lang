package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCheckCompound_AcceptsHyphenParts(t *testing.T) {
	sp := NewMorfologikSpeller("/en/x.dict", 1)
	sp.AddWord("well")
	sp.AddWord("known")
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en", "/en/x.dict", sp)
	// default CheckCompound false → whole "well-known" misspelled
	require.True(t, r.IsMisspelled("well-known"))
	r.SetCheckCompound(true)
	// parts accepted → whole accepted
	require.False(t, r.IsMisspelled("well-known"))
	// one part bad → still misspelled
	require.True(t, r.IsMisspelled("well-knon"))
}

func TestCheckCompound_Match(t *testing.T) {
	sp := NewMorfologikSpeller("/en/x.dict", 1)
	sp.AddWord("well")
	sp.AddWord("known")
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en", "/en/x.dict", sp)
	r.SetCheckCompound(true)
	m, err := r.Match(languagetool.AnalyzePlain("well-known"))
	require.NoError(t, err)
	require.Empty(t, m)
}
