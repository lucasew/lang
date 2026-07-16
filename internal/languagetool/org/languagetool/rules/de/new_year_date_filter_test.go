package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewYearDateFilter(t *testing.T) {
	f := NewNewYearDateFilter()
	jan := true
	y := 2014
	f.ForceJanuary = &jan
	f.ForceYear = &y
	require.True(t, f.ShouldFlag(2013, 3))
	require.False(t, f.ShouldFlag(2013, 12))
	require.False(t, f.ShouldFlag(2014, 3))
	m, err := f.MonthNumber("Januar")
	require.NoError(t, err)
	require.Equal(t, 1, m)
}
