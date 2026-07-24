package rules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testDateCheckFilter() *AbstractDateCheckFilter {
	return &AbstractDateCheckFilter{
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
}

func TestAbstractDateCheckFilter_Mismatch(t *testing.T) {
	f := testDateCheckFilter()
	// 8 November 2003 was a Saturday, not Monday
	m := NewRuleMatch(NewFakeRule("DATE"), nil, 0, 10, "Said {day} but was {realDay}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2003", "month": "November", "day": "8", "weekDay": "Monday",
	})
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "Saturday")
	require.Contains(t, out.GetMessage(), "Monday")
	require.Equal(t, "https://www.timeanddate.com/calendar/?year=2003", out.GetURL())

	// correct weekday — 10 November 2003 was Monday
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2003", "month": "November", "day": "10", "weekDay": "Monday",
	})
	require.Nil(t, ok)
}

func TestAbstractDateCheckFilter_DaySuffixFullMatch(t *testing.T) {
	f := testDateCheckFilter()
	m := NewRuleMatch(NewFakeRule("DATE"), nil, 0, 10, "Said {day} but was {realDay}")
	// "8th" must match DAY_OF_MONTH_PATTERN full match like Java Matcher.matches
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2003", "month": "November", "day": "8th", "weekDay": "Monday",
	})
	require.NotNil(t, out)
	// invent partial match would accept "x8" as day 8; Java does not
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2003", "month": "November", "day": "x8", "weekDay": "Monday",
	}))
}

func TestAbstractDateCheckFilter_MonthTrimSpecial(t *testing.T) {
	// Java StringTools.trimSpecialCharacters strips soft hyphen inside the month token
	f := testDateCheckFilter()
	m := NewRuleMatch(NewFakeRule("DATE"), nil, 0, 10, "Said {day} but was {realDay}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2003", "month": "Novem\u00ADber", "day": "8", "weekDay": "Monday",
	})
	require.NotNil(t, out, "soft-hyphen month must trim like Java trimSpecialCharacters")
}

func TestAbstractDateCheckFilter_InvalidDate(t *testing.T) {
	f := testDateCheckFilter()
	m := NewRuleMatch(NewFakeRule("DATE"), nil, 0, 10, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "32", "weekDay": "Monday",
	}))
}
