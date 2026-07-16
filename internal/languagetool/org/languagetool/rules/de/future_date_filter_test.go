package de

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFutureDateFilter(t *testing.T) {
	f := NewFutureDateFilter()
	require.False(t, f.IsFuture(2000, 1, 1))
	y := time.Now().UTC().Year() + 1
	require.True(t, f.IsFuture(y, 6, 15))
	d, err := ParseDayOfMonth("23.")
	require.NoError(t, err)
	require.Equal(t, 23, d)
}
