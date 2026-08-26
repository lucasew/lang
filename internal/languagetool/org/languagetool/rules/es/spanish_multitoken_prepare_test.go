package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpanishMultitokenSpeller_PrepareLineFiltersPOS(t *testing.T) {
	s := NewSpanishMultitokenSpeller()
	require.NotNil(t, s.PrepareLine)
	require.Equal(t, []string{"casa grande"}, s.PrepareLine("casa grande\tNCMS000"))
	require.Equal(t, []string{"foo bar"}, s.PrepareLine("foo bar;_Latin_"))
	require.Equal(t, []string{"loc"}, s.PrepareLine("loc\tLOC_ADV"))
	require.Equal(t, []string{""}, s.PrepareLine("correr rapido\tVMIP3S0"))
	require.Equal(t, []string{"plain multi"}, s.PrepareLine("plain multi"))
}
