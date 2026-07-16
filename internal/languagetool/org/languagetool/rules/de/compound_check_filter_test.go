package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompoundCheckFilter(t *testing.T) {
	f := NewCompoundCheckFilter()
	require.True(t, f.Accept("Zeit", "Punkt"))
	require.True(t, f.Accept("zeit", "punkt"))
	require.False(t, f.Accept("xyz", "abc"))
}
