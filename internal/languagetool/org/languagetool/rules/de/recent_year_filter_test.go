package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecentYearFilter(t *testing.T) {
	f := NewRecentYearFilter()
	y := 2020
	f.ForceYear = &y
	require.True(t, f.Accept(2019, 5))
	require.True(t, f.Accept(2015, 5))
	require.False(t, f.Accept(2014, 5))
	require.False(t, f.Accept(2020, 5))
}
