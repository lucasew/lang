package fr

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
	require.True(t, f.ShouldFlag(2013, 2))
	require.False(t, f.ShouldFlag(2014, 2))
}
