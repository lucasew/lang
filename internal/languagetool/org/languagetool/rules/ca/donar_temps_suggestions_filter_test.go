package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDonarTempsSuggestionsFilter(t *testing.T) {
	f := NewDonarTempsSuggestionsFilter()
	f.SynthHaver = func(suffix string) string { return "ha" }
	f.SynthTenir = func(postag string) string { return "tinc" }
	got := f.Suggest(DonarTempsInput{
		PronomGenderNumber: "1S",
		VerbPostag:         "VMIP3S00",
		CasingModel:        "em",
	})
	require.Contains(t, got, "hi ha temps")
	require.Contains(t, got, "tinc temps")
}

func TestPronomGenderNumberFromP(t *testing.T) {
	// PP1CS000 → person 1, number S at indices 2 and 4
	require.Equal(t, "1S", PronomGenderNumberFromP("PP1CS000"))
}
