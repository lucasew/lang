package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestVerbSynthesizer_FindVerbGroup(t *testing.T) {
	vtag := "VMIP3S0"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Ell", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("menja", &vtag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("pa", nil, nil)),
	}
	v := NewVerbSynthesizer(tokens)
	require.True(t, v.FindVerbGroup())
	require.Equal(t, 1, v.IFirstVerb)
	require.Equal(t, 1, v.ILastVerb)
	v.SetTarget("menjar", "VMN0000")
	require.True(t, v.HasTarget())
}
