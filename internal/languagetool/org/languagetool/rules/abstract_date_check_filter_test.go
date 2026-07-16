package rules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAbstractDateCheckFilter_Mismatch(t *testing.T) {
	// English: January=1, Monday
	f := &AbstractDateCheckFilter{
		TestMode: true,
		GetDayOfWeekName: func(s string) time.Weekday {
			switch s {
			case "Monday":
				return time.Monday
			case "Tuesday":
				return time.Tuesday
			default:
				return time.Sunday
			}
		},
		FormatDayOfWeek: func(t time.Time) string { return t.Weekday().String() },
		GetMonth: func(s string) int {
			if s == "November" {
				return 11
			}
			return 1
		},
	}
	// 8 November 2003 was a Saturday, not Monday
	m := NewRuleMatch(NewFakeRule("DATE"), nil, 0, 10, "Said {day} but was {realDay}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2003", "month": "November", "day": "8", "weekDay": "Monday",
	})
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "Saturday")
	require.Contains(t, out.GetMessage(), "Monday")

	// correct weekday
	// 10 November 2003 was Monday
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2003", "month": "November", "day": "10", "weekDay": "Monday",
	})
	require.Nil(t, ok)
}
