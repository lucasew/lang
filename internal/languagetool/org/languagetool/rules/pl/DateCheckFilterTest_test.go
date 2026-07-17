package pl

// Twin of DateCheckFilterTest (Polish)
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of DateCheckFilterTest.testGetDayOfWeek
func TestDateCheckFilter_GetDayOfWeek(t *testing.T) {
	f := NewDateCheckFilter()
	d, err := f.GetDayOfWeekJava("poniedziałek")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("niedziela")
	require.NoError(t, err)
	require.Equal(t, 1, d)
	d, err = f.GetDayOfWeekJava("sobota")
	require.NoError(t, err)
	require.Equal(t, 7, d)
}

// Port of DateCheckFilterTest.testMonth
func TestDateCheckFilter_Month(t *testing.T) {
	f := NewDateCheckFilter()
	m, err := f.GetMonth("stycznia")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("III")
	require.NoError(t, err)
	require.Equal(t, 3, m)
}
