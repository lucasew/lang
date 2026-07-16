package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRomanNumeralFilter(t *testing.T) {
	f := NewRomanNumeralFilter()
	require.Equal(t, "I", f.Suggest("1"))
	require.Equal(t, "IV", f.Suggest("4"))
	require.Equal(t, "IX", f.Suggest("9"))
	require.Equal(t, "XLII", f.Suggest("42"))
	require.Equal(t, "MCMXCIX", f.Suggest("1999"))
	require.Equal(t, "MMXXIV", f.Suggest("2024"))
	require.Equal(t, "", f.Suggest("0"))
	require.Equal(t, "", f.Suggest("abc"))
}
