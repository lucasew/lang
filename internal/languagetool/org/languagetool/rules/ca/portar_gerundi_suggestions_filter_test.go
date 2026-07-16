package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortarGerundiSuggestionsFilter(t *testing.T) {
	f := NewPortarGerundiSuggestionsFilter()
	f.SynthHaverParticiple = func(lemma, suffix string) []string {
		return JoinHaverParticiple([]string{"he"}, []string{"fet"})
	}
	f.SynthFinite = func(lemma, suffix string) []string {
		return []string{"faig"}
	}
	got := f.Suggest("VMIP1S00", "fer", "ho", "porto")
	require.Contains(t, got, "ho he fet")
	require.Contains(t, got, "ho faig")
}
