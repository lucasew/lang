package ru

// Twin of DateCheckFilterTest (Russian)
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of DateCheckFilterTest.testGetDayOfWeek (active asserts only).
// Java Calendar: Sunday=1 … Saturday=7
func TestDateCheckFilter_GetDayOfWeek(t *testing.T) {
	f := NewDateCheckFilter()
	// assertThat(filter.getDayOfWeek("пн"), is(2));
	d, err := f.GetDayOfWeekJava("пн")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	// assertThat(filter.getDayOfWeek("пн."), is(2));
	d, err = f.GetDayOfWeekJava("пн.")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	// assertThat(filter.getDayOfWeek("вт"), is(3));
	d, err = f.GetDayOfWeekJava("вт")
	require.NoError(t, err)
	require.Equal(t, 3, d)
	// assertThat(filter.getDayOfWeek("пт"), is(6));
	d, err = f.GetDayOfWeekJava("пт")
	require.NoError(t, err)
	require.Equal(t, 6, d)
}

// Port of DateCheckFilterTest.testMonth
func TestDateCheckFilter_Month(t *testing.T) {
	f := NewDateCheckFilter()
	// assertThat(filter.getMonth("I"), is(1));
	m, err := f.GetMonth("I")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	// assertThat(filter.getMonth("XII"), is(12));
	m, err = f.GetMonth("XII")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	// assertThat(filter.getMonth("декабрь"), is(12));
	m, err = f.GetMonth("декабрь")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	// assertThat(filter.getMonth("Декабрь"), is(12));
	m, err = f.GetMonth("Декабрь")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	// assertThat(filter.getMonth("ДЕКАБРЬ"), is(12));
	m, err = f.GetMonth("ДЕКАБРЬ")
	require.NoError(t, err)
	require.Equal(t, 12, m)
}
