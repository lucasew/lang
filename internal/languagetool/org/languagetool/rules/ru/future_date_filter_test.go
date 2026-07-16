package ru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFutureDateFilter_IsFuture(t *testing.T) {
	f := NewFutureDateFilter()
	require.False(t, f.IsFuture(2000, 1, 1))
	y := time.Now().UTC().Year() + 2
	require.True(t, f.IsFuture(y, 6, 15))
}

func TestParseDayOfMonth(t *testing.T) {
	n, err := ParseDayOfMonth("23.")
	require.NoError(t, err)
	require.Equal(t, 23, n)
}
