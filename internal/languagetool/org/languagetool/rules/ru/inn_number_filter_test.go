package ru

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestINNNumberFilter(t *testing.T) {
	f := NewINNNumberFilter()
	// Known valid 10-digit INN examples (checksum)
	// Compute: use a known-good one. 7830002293 is a classic valid example.
	require.False(t, f.IsInvalid("7830002293"))
	// Flip last digit → invalid
	require.True(t, f.IsInvalid("7830002294"))
	// wrong length → suppress
	require.False(t, f.IsInvalid("123"))
	require.False(t, f.IsInvalid("abc"))
}
