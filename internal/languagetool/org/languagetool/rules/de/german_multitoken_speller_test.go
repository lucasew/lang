package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanMultitokenSpeller_IsException(t *testing.T) {
	s := GermanMultitokenSpeller{}
	require.True(t, s.IsException("Autos", "Auto"))
	require.True(t, s.IsException("foo-", "foo"))
	require.False(t, s.IsException("Haus", "Häuser"))
	require.False(t, s.IsException("Auto", "Autos"))
}
