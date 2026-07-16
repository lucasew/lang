package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYMDNewYearDateFilter(t *testing.T) {
	f := NewYMDNewYearDateFilter()
	jan := true
	y := 2014
	f.newYear.ForceJanuary = &jan
	f.newYear.ForceYear = &y
	ok, err := f.ShouldFlagFromArgs(map[string]string{"date": "2013-03-15", "weekDay": "Fr"})
	require.NoError(t, err)
	require.True(t, ok)
	_, err = f.PrepareArgs(map[string]string{"date": "2013-03-15", "year": "2013"})
	require.Error(t, err)
}
