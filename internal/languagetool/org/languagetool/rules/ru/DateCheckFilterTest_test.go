package ru

// Twin of DateCheckFilterTest (Russian)
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of DateCheckFilterTest.testGetDayOfWeek
func TestDateCheckFilter_GetDayOfWeek(t *testing.T) {
	f := NewDateCheckFilter()
	// Java Calendar: Sunday=1 … Saturday=7
	d, err := f.GetDayOfWeekJava("понедельник")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("пн")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("суббота")
	require.NoError(t, err)
	require.Equal(t, 7, d)
	d, err = f.GetDayOfWeekJava("воскресенье")
	require.NoError(t, err)
	require.Equal(t, 1, d)
}

// Port of DateCheckFilterTest.testMonth
func TestDateCheckFilter_Month(t *testing.T) {
	f := NewDateCheckFilter()
	m, err := f.GetMonth("январь")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("декабрь")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	m, err = f.GetMonth("мая")
	require.NoError(t, err)
	require.Equal(t, 5, m)
}
