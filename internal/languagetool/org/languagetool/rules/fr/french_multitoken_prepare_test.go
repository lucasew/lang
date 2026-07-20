package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin: French.prepareLineForSpeller filters multiwords into MultitokenSpeller.
func TestFrenchMultitokenSpeller_PrepareLineFiltersPOS(t *testing.T) {
	s := NewFrenchMultitokenSpeller()
	require.NotNil(t, s.PrepareLine)
	require.Equal(t, []string{""}, s.PrepareLine("manger bien\tV"))
	require.Equal(t, []string{"Paris ville"}, s.PrepareLine("Paris ville\tZ"))
	require.Equal(t, []string{"bon jour"}, s.PrepareLine("bon jour\tA"))
	require.Equal(t, []string{""}, s.PrepareLine("Ho Chi Minh"))
	require.Equal(t, []string{"plain multi"}, s.PrepareLine("plain multi"))
}
