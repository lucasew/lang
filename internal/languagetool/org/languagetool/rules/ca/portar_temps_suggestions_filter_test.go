package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortarTempsSuggestionsFilter(t *testing.T) {
	f := NewPortarTempsSuggestionsFilter()
	f.SynthFer = func(p string) string { return "fa" }
	got := f.Suggest(PortarTempsInput{
		PortarPostag: "VMIP3S00",
		TimeTokens:   []string{"una", "hora"},
		Kind:         PortarTempsQue,
		CasingModel:  "porta",
	})
	require.Equal(t, "fa una hora que", got)

	f.SynthInfinitiveToFinite = func(lemma, tag string) string { return "treballa" }
	got = f.Suggest(PortarTempsInput{
		PortarPostag:  "VMIP3S00",
		TimeTokens:    []string{"una", "hora"},
		Kind:          PortarTempsGerund,
		NextLemma:     "treballar",
		PronounsAfter: "ho",
	})
	require.Contains(t, got, "fa una hora que")
	require.Contains(t, got, "treballa")
}
