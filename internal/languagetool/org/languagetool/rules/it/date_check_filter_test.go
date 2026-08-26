package it

import (
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Expanded outcome matrix for DateCheckFilter.getDayOfWeek(String).
// Java Calendar Sunday=1 … Saturday=7.
// Prefixes do/lu/ma/me/gi/ve/sa and full Italian names as Java accepts.
func TestDateCheckFilter_GetDayOfWeek_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		// Sunday: startsWith("do") || equals("domenica")
		{"do", 1},
		{"dom", 1},
		{"domenica", 1},
		{"Domenica", 1},
		{"domenicaa", 1}, // prefix
		// Monday: startsWith("lu") || equals("lunedì")
		{"lu", 2},
		{"lun", 2},
		{"lunedì", 2},
		{"Lunedì", 2},
		// Tuesday: startsWith("ma") || equals("martedì")
		{"ma", 3},
		{"mar", 3},
		{"martedì", 3},
		{"Martedì", 3},
		// Wednesday: startsWith("me") || equals("mercoledì")
		{"me", 4},
		{"mer", 4},
		{"mercoledì", 4},
		// Thursday: startsWith("gi") || equals("giovedì")
		{"gi", 5},
		{"gio", 5},
		{"giovedì", 5},
		// Friday: startsWith("ve") || equals("venerdì")
		{"ve", 6},
		{"ven", 6},
		{"venerdì", 6},
		// Saturday: startsWith("sa") || equals("sabato")
		{"sa", 7},
		{"sab", 7},
		{"sabato", 7},
		{"Sabato", 7},
	}
	for _, tc := range cases {
		d, err := f.GetDayOfWeekJava(tc.in)
		require.NoError(t, err, "day %q", tc.in)
		require.Equal(t, tc.want, d, "day %q", tc.in)
	}
}

func TestDateCheckFilter_GetDayOfWeek_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	// English / unknown must fail (no invent).
	for _, in := range []string{"", "xyz", "monday", "Mon", "thursday", "Sunday", "friday"} {
		_, err := f.GetDayOfWeekJava(in)
		require.Error(t, err, "expected error for day %q", in)
		require.Contains(t, err.Error(), "could not find day of week for '"+in+"'")
	}
}

// getDayOfWeek(Calendar): Locale.UK LONG English → Italian names.
func TestDateCheckFilter_GetDayOfWeek_CalendarFormat(t *testing.T) {
	f := NewDateCheckFilter()
	// 2022-03-20..26: Sunday … Saturday
	require.Equal(t, "domenica", f.GetDayOfWeekName(2022, 3, 20))
	require.Equal(t, "lunedì", f.GetDayOfWeekName(2022, 3, 21))
	require.Equal(t, "martedì", f.GetDayOfWeekName(2022, 3, 22))
	require.Equal(t, "mercoledì", f.GetDayOfWeekName(2022, 3, 23))
	require.Equal(t, "giovedì", f.GetDayOfWeekName(2022, 3, 24))
	require.Equal(t, "venerdì", f.GetDayOfWeekName(2022, 3, 25))
	require.Equal(t, "sabato", f.GetDayOfWeekName(2022, 3, 26))
	// Known dates used in AcceptRuleMatch twins
	require.Equal(t, "mercoledì", f.GetDayOfWeekName(2014, 8, 27))
	require.Equal(t, "martedì", f.GetDayOfWeekName(2014, 8, 26))
	require.Equal(t, "lunedì", f.GetDayOfWeekName(2024, 1, 1))
}

// Month prefixes gen…dic exactly as Java.
func TestDateCheckFilter_GetMonth_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		{"gen", 1}, {"gennaio", 1}, {"Gennaio", 1},
		{"feb", 2}, {"febbraio", 2},
		{"mar", 3}, {"marzo", 3},
		{"apr", 4}, {"aprile", 4},
		{"mag", 5}, {"maggio", 5},
		{"giu", 6}, {"giugno", 6},
		{"lug", 7}, {"luglio", 7},
		{"ago", 8}, {"agosto", 8}, {"Agosto", 8},
		{"set", 9}, {"settembre", 9},
		{"ott", 10}, {"ottobre", 10},
		{"nov", 11}, {"novembre", 11},
		{"dic", 12}, {"dicembre", 12},
	}
	for _, tc := range cases {
		got, err := f.GetMonth(tc.in)
		require.NoError(t, err, "month %q", tc.in)
		require.Equal(t, tc.want, got, "month %q", tc.in)
	}
}

func TestDateCheckFilter_GetMonth_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	// English / non-IT prefixes fail. Note "may" does not start with "mag".
	// "jan" does not start with "gen". "aug" is accepted (ago prefix) — Java prefix.
	for _, in := range []string{"", "xyz", "january", "may", "June", "december"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "expected error for month %q", in)
		require.Contains(t, err.Error(), "could not find month '"+in+"'")
	}
	// "aug" startsWith("ago") is false — Java only "ago"
	_, err := f.GetMonth("aug")
	require.Error(t, err)
	// "ago" is the real Italian prefix
	m, err := f.GetMonth("ago")
	require.NoError(t, err)
	require.Equal(t, 8, m)
}

// Helper surface still matches filter for localization primitives.
func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	wd, err := h.GetDayOfWeek("lunedì")
	require.NoError(t, err)
	require.Equal(t, time.Monday, wd)
	m, err := h.GetMonth("agosto")
	require.NoError(t, err)
	require.Equal(t, time.August, m)
	// prefixes
	wd, err = h.GetDayOfWeek("ma")
	require.NoError(t, err)
	require.Equal(t, time.Tuesday, wd)
	m, err = h.GetMonth("gen")
	require.NoError(t, err)
	require.Equal(t, time.January, m)
	// unknown
	_, err = h.GetDayOfWeek("monday")
	require.Error(t, err)
	_, err = h.GetMonth("january")
	require.Error(t, err)
}

func TestDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewDateCheckFilter()
	// 27 agosto 2014 was Wednesday (mercoledì), not martedì
	m := rules.NewRuleMatch(nil, nil, 0, 20, "Giorno {realDay}, non {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "agosto", "day": "27", "weekDay": "martedì",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "mercoledì")
	require.Contains(t, out.GetMessage(), "martedì")

	// 26 agosto 2014 was Tuesday = martedì
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "agosto", "day": "26", "weekDay": "martedì",
	}, 0, nil, nil)
	require.Nil(t, ok)

	// Numeric month twin: 2022-03-12 Saturday correct / Sunday wrong
	match := rules.NewRuleMatch(nil, nil, 0, 10, "{realDay}/{day}")
	require.Nil(t, f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "sabato",
	}, -1, nil, nil))
	bad := f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "domenica",
	}, -1, nil, nil)
	require.NotNil(t, bad)
	require.Contains(t, bad.GetMessage(), "sabato")
	require.Contains(t, bad.GetMessage(), "domenica")

	// Prefix weekday: "lu" vs real mercoledì on 2014-08-27
	prefixBad := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "ago", "day": "27", "weekDay": "lu",
	}, 0, nil, nil)
	require.NotNil(t, prefixBad)
	require.Contains(t, prefixBad.GetMessage(), "mercoledì")
	require.Contains(t, prefixBad.GetMessage(), "lunedì") // {day} is FormatDayOfWeek of claimed weekday

	// Year optional (grammar rule without year) uses TestMode → 2014
	f.TestMode = true
	out = f.AcceptRuleMatch(m, map[string]string{
		"month": "agosto", "day": "27", "weekDay": "lunedì",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "mercoledì")
}

func TestDateCheckFilter_UnknownWeekdaySkips(t *testing.T) {
	// Java RuntimeException from getDayOfWeek → logger.warn + return null (not crash)
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "notaday",
	}, -1, nil, nil)
	require.Nil(t, out)
}

func TestDateCheckFilter_UnknownMonthSkips(t *testing.T) {
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2022", "month": "notamonth", "day": "12", "weekDay": "sabato",
	}, -1, nil, nil)
	require.Nil(t, out)
	// English "january" must skip (RuntimeException path), not invent January
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "january", "day": "27", "weekDay": "lunedì",
	}, -1, nil, nil)
	require.Nil(t, out)
}

func TestDateCheckFilter_MissingWeekDayPanics(t *testing.T) {
	f := NewDateCheckFilter()
	match := rules.NewRuleMatch(nil, nil, 0, 10, "message")
	// Java: IllegalArgumentException when weekDay missing
	require.Panics(t, func() {
		f.AcceptRuleMatch(match, map[string]string{
			"year": "2022", "month": "3", "day": "12",
		}, -1, nil, nil)
	})
}

func TestITDateCheckFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.it.DateCheckFilter"))
}
