package rules

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFutureDateFilterCore(t *testing.T) {
	f := &FutureDateFilterCore{
		Now: func() time.Time { return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.Local) },
	}
	require.True(t, f.IsFuture(2014, 6, 1))
	require.False(t, f.IsFuture(2013, 6, 1))
	require.False(t, f.IsFuture(2014, 1, 1), "not strictly after")
	require.True(t, f.AcceptFromArgs(map[string]string{"year": "2015", "month": "3", "day": "10"}))
	require.False(t, f.AcceptFromArgs(map[string]string{"year": "2013", "month": "3", "day": "10"}))
	// invalid date (Java setLenient(false) → after() throws → null)
	require.False(t, f.AcceptFromArgs(map[string]string{"year": "2015", "month": "2", "day": "30"}))
	require.False(t, f.AcceptFromArgs(map[string]string{"year": "2015", "month": "2", "day": "32"}))
}

func TestFutureDateFilterCore_LocalizedMonth(t *testing.T) {
	f := &FutureDateFilterCore{
		Now: func() time.Time { return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.Local) },
		GetMonth: func(localized string) (int, error) {
			if localized == "March" {
				return 3, nil
			}
			return 0, errors.New("unknown month")
		},
	}
	require.True(t, f.AcceptFromArgs(map[string]string{"year": "2015", "month": "March", "day": "10"}))
	// numeric month still works without GetMonth path for digits
	require.True(t, f.AcceptFromArgs(map[string]string{"year": "2015", "month": "3", "day": "10"}))
}

func TestParseDayOfMonthArg(t *testing.T) {
	n, err := ParseDayOfMonthArg("22nd", nil)
	require.NoError(t, err)
	require.Equal(t, 22, n)
	// full Matcher.matches — invent leading-digit-only would still work, but mid invent "x22" must fail
	n, err = ParseDayOfMonthArg("x22", nil)
	require.NoError(t, err)
	require.Equal(t, 0, n, "no full match → Java getDayOfMonth default 0")
	// named day via hook
	n, err = ParseDayOfMonthArg("first", func(s string) (int, error) {
		if s == "first" {
			return 1, nil
		}
		return 0, nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	// day required; missing panics
	f := &FutureDateFilterCore{
		Now: func() time.Time { return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.Local) },
	}
	require.Panics(t, func() {
		f.AcceptFromArgs(map[string]string{"year": "2015", "month": "3"})
	})
	require.Panics(t, func() {
		f.AcceptFromArgs(map[string]string{"year": "2015", "day": "10"})
	})
	require.Panics(t, func() {
		f.AcceptFromArgs(map[string]string{"month": "3", "day": "10"})
	})
}

func TestFutureDateFilterCore_MonthMinusOneBoundary(t *testing.T) {
	// Java getMonthFromArguments returns month-1 for Calendar; January=0 … December=11.
	// Go keeps 1-based; January=1 must still be future vs 2014-01-01 when day>1.
	f := &FutureDateFilterCore{
		Now: func() time.Time { return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.Local) },
	}
	require.True(t, f.AcceptFromArgs(map[string]string{"year": "2014", "month": "1", "day": "2"}))
	require.True(t, f.AcceptFromArgs(map[string]string{"year": "2014", "month": "12", "day": "31"}))
	// month "01" isNumeric → 1
	require.True(t, f.AcceptFromArgs(map[string]string{"year": "2014", "month": "01", "day": "2"}))
}
