package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBrazilianToponymMap(t *testing.T) {
	m := LoadBrazilianToponymMap()
	require.True(t, m.IsValidToponym("São Paulo"))
	require.True(t, m.IsValidToponym("Venho do Rio de Janeiro")) // suffix match
	require.False(t, m.IsValidToponym("Narnia"))
	require.True(t, m.IsToponymInState("são paulo", "SP"))
}

func TestBrazilianToponymFilter(t *testing.T) {
	f := NewBrazilianToponymFilter()
	require.Equal(t, "–SP", f.Suggest("São Paulo", "-", "SP"))
	require.Equal(t, "", f.Suggest("São Paulo", "–SP", "SP"))
	require.Equal(t, "", f.Suggest("Narnia", "-", "XX"))
}
