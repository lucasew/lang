package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/DateCheckFilterTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of DateCheckFilterTest.testGetDayOfWeek
func TestDateCheckFilter_GetDayOfWeek(t *testing.T) {
	f := NewDateCheckFilter()
	// Java Calendar: Sunday=1 … Saturday=7
	d, err := f.GetDayOfWeekJava("Sun")
	require.NoError(t, err)
	require.Equal(t, 1, d)
	d, err = f.GetDayOfWeekJava("Mon")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("mon")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("Monday")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("Friday")
	require.NoError(t, err)
	require.Equal(t, 6, d)
	d, err = f.GetDayOfWeekJava("Saturday")
	require.NoError(t, err)
	require.Equal(t, 7, d)
}

// Port of DateCheckFilterTest.testMonth
func TestDateCheckFilter_Month(t *testing.T) {
	f := NewDateCheckFilter()
	m, err := f.GetMonth("jan")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("december")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	m, err = f.GetMonth("DECEMBER")
	require.NoError(t, err)
	require.Equal(t, 12, m)
}
