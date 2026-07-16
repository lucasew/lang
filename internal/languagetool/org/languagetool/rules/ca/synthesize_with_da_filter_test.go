package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenderNumberFromPOS(t *testing.T) {
	require.Equal(t, "MS", GenderNumberFromPOS("NCMS000"))
	require.Equal(t, "FS", GenderNumberFromPOS("NCFS000"))
	require.Equal(t, "MP", GenderNumberFromPOS("NCMP000"))
}

func TestSynthesizeWithDAFilter_Prefixed(t *testing.T) {
	f := NewSynthesizeWithDAFilter()
	require.Equal(t, "l'amic", f.PrefixedSuggestion("amic", "MS", ""))
	require.Equal(t, "de la casa", f.PrefixedSuggestion("casa", "FS", "de"))
}
