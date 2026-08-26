package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVerbInflections(t *testing.T) {
	got := GetVerbInflections([]string{"verb:imperf:pres:s:3"})
	require.NotEmpty(t, got)
	require.Equal(t, "s", got[0].Plural)
	require.Equal(t, "3", got[0].Person)

	got = GetVerbInflections([]string{"verb:inf"})
	require.Equal(t, "i", got[0].Gender)
}

func TestVerbInflectionsOverlap(t *testing.T) {
	// verb s:3 and noun m:v_naz with person absent — gender match via s vs m?
	// NewVerbInflection("s") → Plural s, Gender ""
	// NewVerbInflection("m") → Gender m, Plural s
	// Equals: gender empty on one side → true if person ok
	require.True(t, VerbInflectionsOverlap(
		[]string{"verb:imperf:pres:s:3"},
		[]string{"noun:inanim:m:v_naz"},
	))
}
