package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanMultitokenSpeller_PrepareLineFiltersPOS(t *testing.T) {
	s := NewCatalanMultitokenSpeller()
	require.NotNil(t, s.PrepareLine)
	require.Equal(t, []string{"casa gran"}, s.PrepareLine("casa gran\tNCMS000"))
	require.Equal(t, []string{"foo bar"}, s.PrepareLine("foo bar;_Latin_"))
	require.Equal(t, []string{""}, s.PrepareLine("córrer ràpid\tVMIP3S0"))
	require.Equal(t, []string{""}, s.PrepareLine("Banco Santander"))
	require.Equal(t, []string{"plain multi"}, s.PrepareLine("plain multi"))
}
