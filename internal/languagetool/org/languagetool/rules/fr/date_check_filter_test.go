package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter(t *testing.T) {
	f := NewDateCheckFilter()
	d, err := f.GetDayOfWeekJava("lundi")
	require.NoError(t, err)
	require.Equal(t, 2, d) // Java Monday=2
	m, err := f.GetMonth("janvier")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	require.Equal(t, "vendredi", f.GetDayOfWeekName(2014, 8, 29))
}
