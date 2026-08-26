package rules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewYearDateFilterCore(t *testing.T) {
	jan := true
	y := 2014
	f := &NewYearDateFilterCore{ForceJanuary: &jan, ForceYear: &y}
	require.True(t, f.ShouldFlag(2013, 3))
	require.False(t, f.ShouldFlag(2013, 12)) // December skipped (Java monthFromText != 11)
	require.False(t, f.ShouldFlag(2014, 3))  // wrong year (not currentYear-1)
	// Java always validates day with setLenient(false)
	msg := f.AcceptFromArgs(map[string]string{"year": "2013", "month": "5", "day": "10"}, "was {year} now {realYear}")
	require.Equal(t, "was 2013 now 2014", msg)
	// invalid day → suppress (IllegalArgumentException → null)
	require.Empty(t, f.AcceptFromArgs(map[string]string{"year": "2013", "month": "2", "day": "30"}, "was {year}"))
	// soft hyphen stripped before day pattern (Java)
	require.Equal(t, "was 2013 now 2014", f.AcceptFromArgs(
		map[string]string{"year": "2013", "month": "5", "day": "1\u00AD0"}, "was {year} now {realYear}"))
	// soft hyphen stripped before month parse (Java getMonthFromArguments)
	require.Equal(t, "was 2013 now 2014", f.AcceptFromArgs(
		map[string]string{"year": "2013", "month": "0\u00AD5", "day": "10"}, "was {year} now {realYear}"))
	// DAY_OF_MONTH_PATTERN: "22nd" → 22
	require.Equal(t, "was 2013 now 2014", f.AcceptFromArgs(
		map[string]string{"year": "2013", "month": "5", "day": "22nd"}, "was {year} now {realYear}"))
	// spelled-out day with default getDayOfMonth → 0 → invalid calendar → null
	require.Empty(t, f.AcceptFromArgs(
		map[string]string{"year": "2013", "month": "5", "day": "first"}, "was {year}"))
}

func TestNewYearDateFilterCore_NotJanuary(t *testing.T) {
	jan := false
	y := 2014
	f := &NewYearDateFilterCore{ForceJanuary: &jan, ForceYear: &y}
	require.False(t, f.ShouldFlag(2013, 3))
	require.Empty(t, f.AcceptFromArgs(
		map[string]string{"year": "2013", "month": "3", "day": "1"}, "y {year}"))
}

func TestNewYearDateFilterCore_TestHack(t *testing.T) {
	// Under go test, IsTest() → January + year 2014 without Force*
	f := &NewYearDateFilterCore{}
	require.True(t, f.isJanuary())
	require.Equal(t, 2014, f.currentYear())
	require.True(t, f.ShouldFlag(2013, 6))
}

func TestNewYearDateFilterCore_GetCalendar(t *testing.T) {
	// Force* nil and GetCalendar providing non-January → no flag (when not IsTest).
	// Under go test IsTest wins over GetCalendar for isJanuary/year — same as Java TestHack.
	cal := time.Date(2020, time.March, 15, 0, 0, 0, 0, time.UTC)
	f := &NewYearDateFilterCore{
		GetCalendar: func() time.Time { return cal },
	}
	// IsTest() still true under go test
	require.True(t, f.isJanuary())
	require.Equal(t, 2014, f.currentYear())
}

func TestNewYearDateFilterCore_MissingKeys(t *testing.T) {
	jan := true
	y := 2014
	f := &NewYearDateFilterCore{ForceJanuary: &jan, ForceYear: &y}
	require.Panics(t, func() {
		f.AcceptFromArgs(map[string]string{"month": "1", "day": "1"}, "m")
	})
	require.Panics(t, func() {
		f.AcceptFromArgs(map[string]string{"year": "2013", "day": "1"}, "m")
	})
	require.Panics(t, func() {
		f.AcceptFromArgs(map[string]string{"year": "2013", "month": "1"}, "m")
	})
}

func TestNewYearDateFilterCore_GetMonthLocalized(t *testing.T) {
	jan := true
	y := 2014
	f := &NewYearDateFilterCore{
		ForceJanuary: &jan,
		ForceYear:    &y,
		GetMonth: func(s string) (int, error) {
			if s == "March" {
				return 3, nil
			}
			return 0, strconvErr(s)
		},
	}
	require.Equal(t, "was 2013 now 2014", f.AcceptFromArgs(
		map[string]string{"year": "2013", "month": "March", "day": "15"}, "was {year} now {realYear}"))
	// soft hyphen on localized month
	require.Equal(t, "was 2013 now 2014", f.AcceptFromArgs(
		map[string]string{"year": "2013", "month": "Mar\u00ADch", "day": "15"}, "was {year} now {realYear}"))
}

func strconvErr(s string) error {
	return &monthErr{s}
}

type monthErr struct{ s string }

func (e *monthErr) Error() string { return "unknown month " + e.s }
