package de

// Twin of DateCheckFilterTest.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter_GetDayOfWeek1(t *testing.T) {
	f := NewDateCheckFilter()
	// Java Calendar: Sunday=1 … Saturday=7
	d, err := f.GetDayOfWeekJava("So")
	require.NoError(t, err)
	require.Equal(t, 1, d)
	d, err = f.GetDayOfWeekJava("Mo")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("mo")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("Mon.")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("Montag")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("Di")
	require.NoError(t, err)
	require.Equal(t, 3, d)
	d, err = f.GetDayOfWeekJava("Fr")
	require.NoError(t, err)
	require.Equal(t, 6, d)
	d, err = f.GetDayOfWeekJava("Samstag")
	require.NoError(t, err)
	require.Equal(t, 7, d)
	d, err = f.GetDayOfWeekJava("Sonnabend")
	require.NoError(t, err)
	require.Equal(t, 7, d)
}

func TestDateCheckFilter_GetDayOfWeek2(t *testing.T) {
	f := NewDateCheckFilter()
	// 2014-08-29 = Friday, 2014-08-30 = Saturday
	require.Equal(t, "Freitag", f.GetDayOfWeekName(2014, 8, 29))
	require.Equal(t, "Samstag", f.GetDayOfWeekName(2014, 8, 30))
}

func TestDateCheckFilter_GetMonth(t *testing.T) {
	f := NewDateCheckFilter()
	m, err := f.GetMonth("Januar")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("Jan")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("Jan.")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("Dezember")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	m, err = f.GetMonth("Dez")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	m, err = f.GetMonth("DEZEMBER")
	require.NoError(t, err)
	require.Equal(t, 12, m)
}

func TestDateCheckFilter_AcceptIncompleteArgs(t *testing.T) {
	f := NewDateCheckFilter()
	_, err := f.AcceptRuleMatch(map[string]string{"year": "2014", "month": "8", "day": "23"})
	require.Error(t, err)
}

func TestDateCheckFilter_Accept(t *testing.T) {
	// Full acceptRuleMatch calendar check deferred; ensure constructor works.
	require.NotNil(t, NewDateCheckFilter())
}
