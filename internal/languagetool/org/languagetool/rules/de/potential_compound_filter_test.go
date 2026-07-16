package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPotentialCompoundFilter(t *testing.T) {
	f := NewPotentialCompoundFilter()
	s := f.Suggestions("Haus", "tür")
	require.Contains(t, s, "Haustür")
	f.JoinedIsValid = func(string) bool { return false }
	s2 := f.Suggestions("Haus", "Tür")
	require.Equal(t, []string{"Haus-Tür"}, s2)
}
