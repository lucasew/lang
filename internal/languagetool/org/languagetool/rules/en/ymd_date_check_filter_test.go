package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestYMDDateCheckFilter_ParseDateViaHelper(t *testing.T) {
	// Sector twin: parseDate outcomes used by YMDDateCheckFilter (mutates args).
	h := rules.NewYMDDateHelper()
	args := map[string]string{"date": "1999-12-31", "weekDay": "6"}
	out, err := h.ParseDate(args)
	require.NoError(t, err)
	require.Equal(t, "1999", out["year"])
	require.Equal(t, "12", out["month"])
	require.Equal(t, "31", out["day"])
	require.Equal(t, "1999", args["year"])
}

func TestYMDDateCheckFilter_AcceptRuleMatch_WrongWeekday(t *testing.T) {
	// 2014-08-23 is Saturday; claiming Sunday keeps the match (delegates to DateCheckFilter).
	f := NewYMDDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "wrong {realDay} not {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"date":    "2014-08-23",
		"weekDay": "Sunday",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "Saturday")
}

func TestYMDDateCheckFilter_AcceptRuleMatch_CorrectWeekday(t *testing.T) {
	f := NewYMDDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"date":    "2014-08-23",
		"weekDay": "Saturday",
	}, 0, nil, nil)
	require.Nil(t, out)
}

func TestYMDDateCheckFilter_WeekdayMismatchMatrix(t *testing.T) {
	// Same calendar date as DateCheckFilter tests; YMD path via date=yyyy-mm-dd.
	f := NewYMDDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "real={realDay} claimed={day}")
	// 2014-08-29 is Friday
	cases := []struct {
		weekDay string
		keep    bool
	}{
		{"Friday", false},
		{"Monday", true},
		{"Saturday", true},
		{"Sunday", true},
	}
	for _, tc := range cases {
		t.Run(tc.weekDay, func(t *testing.T) {
			out := f.AcceptRuleMatch(m, map[string]string{
				"date":    "2014-08-29",
				"weekDay": tc.weekDay,
			}, 0, nil, nil)
			if tc.keep {
				require.NotNil(t, out, "expected keep match for claimed %s", tc.weekDay)
				require.Contains(t, out.GetMessage(), "Friday")
			} else {
				require.Nil(t, out, "expected drop match when weekday correct")
			}
		})
	}
}

func TestYMDDateCheckFilter_RejectsYearMonthDayKeys(t *testing.T) {
	f := NewYMDDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 1, "msg")
	require.PanicsWithValue(t, "Set only 'weekDay' and 'date' for YMDDateCheckFilter", func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "2014-08-23", "year": "2014", "weekDay": "Saturday"}, 0, nil, nil)
	})
	require.PanicsWithValue(t, "Set only 'weekDay' and 'date' for YMDDateCheckFilter", func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "2014-08-23", "month": "8", "weekDay": "Saturday"}, 0, nil, nil)
	})
	require.PanicsWithValue(t, "Set only 'weekDay' and 'date' for YMDDateCheckFilter", func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "2014-08-23", "day": "23", "weekDay": "Saturday"}, 0, nil, nil)
	})
}

func TestYMDDateCheckFilter_MissingDatePanics(t *testing.T) {
	f := NewYMDDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 1, "msg")
	require.PanicsWithValue(t, "Missing key 'date'", func() {
		f.AcceptRuleMatch(m, map[string]string{"weekDay": "Saturday"}, 0, nil, nil)
	})
}

func TestYMDDateCheckFilter_BadDateFormatPanics(t *testing.T) {
	f := NewYMDDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 1, "msg")
	// Java: parts.length != 3 → RuntimeException (e.g. missing day segment).
	require.PanicsWithValue(t, "Expected date in format 'yyyy-mm-dd': '2014-08'", func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "2014-08", "weekDay": "Saturday"}, 0, nil, nil)
	})
}

func TestYMDDateCheckFilter_MutatesArgsForSuper(t *testing.T) {
	// Java parseDate mutates the same map so super sees year/month/day.
	f := NewYMDDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	args := map[string]string{"date": "2014-08-23", "weekDay": "Saturday"}
	_ = f.AcceptRuleMatch(m, args, 0, nil, nil)
	require.Equal(t, "2014", args["year"])
	require.Equal(t, "08", args["month"])
	require.Equal(t, "23", args["day"])
}

func TestYMDDateCheckFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.en.YMDDateCheckFilter"))
}
