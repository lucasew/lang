package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveUnknownCompoundsFilter(t *testing.T) {
	f := NewRemoveUnknownCompoundsFilter()
	require.True(t, f.Accept("Haus", "Tür"))
	f.IsMisspelled = func(w string) bool { return w == "Hausxyz" }
	require.False(t, f.Accept("Haus", "Xyz"))
}
