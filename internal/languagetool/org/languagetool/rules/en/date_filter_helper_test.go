package en

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	d, err := h.GetDayOfWeek("Monday")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	// Abbreviation / trailing period (trimSpecialCharacters keeps '.' for prefix match)
	d, err = h.GetDayOfWeek("Mon.")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	d, err = h.GetDayOfWeek("Fr")
	require.NoError(t, err)
	require.Equal(t, time.Friday, d)
	m, err := h.GetMonth("September")
	require.NoError(t, err)
	require.Equal(t, time.September, m)
	m, err = h.GetMonth("dec")
	require.NoError(t, err)
	require.Equal(t, time.December, m)
	// Locale.UK LONG weekday names
	require.Equal(t, "Friday", h.GetDayOfWeekName(time.Date(2014, 8, 29, 0, 0, 0, 0, time.UTC)))
	require.Equal(t, "Saturday", h.GetDayOfWeekName(time.Date(2014, 8, 23, 0, 0, 0, 0, time.UTC)))
}

func TestDateFilterHelper_TrimSpecialCharacters(t *testing.T) {
	h := NewDateFilterHelper()
	// Java StringTools.trimSpecialCharacters strips soft hyphen, keeps letters
	m, err := h.GetMonth("Sep\u00ADtember")
	require.NoError(t, err)
	require.Equal(t, time.September, m)
	d, err := h.GetDayOfWeek("Mon\u00ADday")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
}
