package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter(t *testing.T) {
	f := NewDateCheckFilter()
	d, err := f.GetDayOfWeekJava("lunes")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	m, err := f.GetMonth("enero")
	require.NoError(t, err)
	require.Equal(t, 1, m)
}
