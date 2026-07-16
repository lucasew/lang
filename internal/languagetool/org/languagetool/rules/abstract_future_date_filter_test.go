package rules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFutureDateFilterCore(t *testing.T) {
	f := &FutureDateFilterCore{
		Now: func() time.Time { return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC) },
	}
	require.True(t, f.IsFuture(2014, 6, 1))
	require.False(t, f.IsFuture(2013, 6, 1))
	require.True(t, f.AcceptFromArgs(map[string]string{"year": "2015", "month": "3", "day": "10"}))
	require.False(t, f.AcceptFromArgs(map[string]string{"year": "2013", "month": "3", "day": "10"}))
	// invalid date
	require.False(t, f.AcceptFromArgs(map[string]string{"year": "2015", "month": "2", "day": "30"}))
}

func TestParseDayOfMonthArg(t *testing.T) {
	n, err := ParseDayOfMonthArg("22nd", nil)
	require.NoError(t, err)
	require.Equal(t, 22, n)
}
