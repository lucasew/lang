package eo

import (
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Expanded outcome matrix for DateCheckFilter.getDayOfWeek(String).
// Java Calendar Sunday=1 … Saturday=7.
func TestDateCheckFilter_GetDayOfWeek_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		// Sunday: startsWith("dim")
		{"dim", 1},
		{"dimanĉo", 1},
		{"Dimanĉo", 1},
		{"dimancon", 1},
		// Monday: startsWith("lun")
		{"lun", 2},
		{"lundo", 2},
		{"Lundo", 2},
		// Tuesday: startsWith("mar")
		{"mar", 3},
		{"mardo", 3},
		// Wednesday: startsWith("mer")
		{"mer", 4},
		{"merkredo", 4},
		// Thursday: ĵaŭ / jau / jhau / jxau (Java order)
		{"ĵaŭ", 5},
		{"ĵaŭdo", 5},
		{"jau", 5},
		{"jaudo", 5},
		{"jhau", 5},
		{"jhaudo", 5},
		{"jxau", 5},
		{"jxaud", 5},
		// Friday: startsWith("ven")
		{"ven", 6},
		{"vendredo", 6},
		// Saturday: startsWith("sab")
		{"sab", 7},
		{"sabato", 7},
	}
	for _, tc := range cases {
		d, err := f.GetDayOfWeekJava(tc.in)
		require.NoError(t, err, "day %q", tc.in)
		require.Equal(t, tc.want, d, "day %q", tc.in)
	}
}

func TestDateCheckFilter_GetDayOfWeek_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	for _, in := range []string{"", "xyz", "monday", "Mon", "thursday", "jaŭdo"} {
		// "jaŭdo" uses ŭ not u — Java startsWith("jau") is plain u; startsWith("ĵaŭ") needs ĵ.
		// Bug-for-bug: "jaŭdo" fails (format name is display-only).
		_, err := f.GetDayOfWeekJava(in)
		require.Error(t, err, "expected error for day %q", in)
		require.Contains(t, err.Error(), "could not find day of week for '"+in+"'")
	}
}

// getDayOfWeek(Calendar): Locale.UK LONG English → EO names.
func TestDateCheckFilter_GetDayOfWeek_CalendarFormat(t *testing.T) {
	f := NewDateCheckFilter()
	// 2022-03-20..26: Sunday … Saturday
	require.Equal(t, "dimanĉo", f.GetDayOfWeekName(2022, 3, 20))
	require.Equal(t, "lundo", f.GetDayOfWeekName(2022, 3, 21))
	require.Equal(t, "mardo", f.GetDayOfWeekName(2022, 3, 22))
	require.Equal(t, "merkredo", f.GetDayOfWeekName(2022, 3, 23))
	require.Equal(t, "jaŭdo", f.GetDayOfWeekName(2022, 3, 24))
	require.Equal(t, "vendredo", f.GetDayOfWeekName(2022, 3, 25))
	require.Equal(t, "sabato", f.GetDayOfWeekName(2022, 3, 26))
	// Known date used in AcceptRuleMatch twins
	require.Equal(t, "merkredo", f.GetDayOfWeekName(2014, 8, 27))
	require.Equal(t, "mardo", f.GetDayOfWeekName(2014, 8, 26))
	require.Equal(t, "lundo", f.GetDayOfWeekName(2024, 1, 1))
}

// Month prefixes jan…dec; August only aŭg (not invent aug).
func TestDateCheckFilter_GetMonth_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		{"jan", 1}, {"januaro", 1}, {"Januaro", 1},
		{"feb", 2}, {"februaro", 2},
		{"mar", 3}, {"marto", 3},
		{"apr", 4}, {"aprilo", 4},
		{"maj", 5}, {"majo", 5},
		{"jun", 6}, {"junio", 6},
		{"jul", 7}, {"julio", 7},
		{"aŭg", 8}, {"aŭgusto", 8}, {"Aŭgusto", 8},
		{"sep", 9}, {"septembro", 9},
		{"okt", 10}, {"oktobro", 10},
		{"nov", 11}, {"novembro", 11},
		{"dec", 12}, {"decembro", 12},
	}
	for _, tc := range cases {
		got, err := f.GetMonth(tc.in)
		require.NoError(t, err, "month %q", tc.in)
		require.Equal(t, tc.want, got, "month %q", tc.in)
	}
}

func TestDateCheckFilter_GetMonth_NoAugInvent(t *testing.T) {
	// Java only mon.startsWith("aŭg") — Latin "aug" must fail.
	f := NewDateCheckFilter()
	for _, in := range []string{"aug", "august", "August", "AUG"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "invented Latin form %q must fail", in)
		require.Contains(t, err.Error(), "could not find month '"+in+"'")
	}
}

func TestDateCheckFilter_GetMonth_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	// "january" startsWith("jan") → accepted (Java prefix, not invent rejection)
	for _, in := range []string{"", "xyz", "mai", "august", "may"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "expected error for month %q", in)
	}
}

// Day-of-month: final -n strip; dek/dudek/tridek; hyphen; unua…naŭa variants.
func TestDateCheckFilter_GetDayOfMonth_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		// base ordinals
		{"unua", 1},
		{"dua", 2},
		{"tria", 3},
		{"kvara", 4},
		{"kvina", 5},
		{"sesa", 6},
		{"sepa", 7},
		{"oka", 8},
		{"naŭa", 9},
		{"nauxa", 9},
		{"naua", 9},
		// accusative final -n
		{"unuan", 1},
		{"trian", 3},
		{"naŭan", 9},
		{"nauxan", 9},
		{"nauan", 9},
		// tens: dek (10–19)
		{"dek", 10},
		{"dekunua", 11},
		{"dek-unua", 11},
		{"dekdua", 12},
		{"dektria", 13},
		{"deksesa", 16},
		{"deknaŭa", 19},
		{"dek-naŭa", 19},
		{"deknauxa", 19},
		// tens: dudek (20–29)
		{"dudek", 20},
		{"dudek-tria", 23},
		{"dudek-trian", 23}, // -n strip before tens parse in Java? endsWith n first
		{"dudekunua", 21},
		{"dudek-naŭa", 29},
		// tens: tridek (30–31)
		{"tridek", 30},
		{"tridek-unua", 31},
		{"tridekunua", 31},
		// case
		{"Tria", 3},
		{"DEKsesa", 16},
		// unknown ordinal residue → tens only or 0
		{"", 0},
		{"xyz", 0},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, f.GetDayOfMonth(tc.in), "dayOfMonth %q", tc.in)
	}
}

// Helper surface still matches filter for localization primitives.
func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	wd, err := h.GetDayOfWeek("lundo")
	require.NoError(t, err)
	require.Equal(t, time.Monday, wd)
	m, err := h.GetMonth("aŭgusto")
	require.NoError(t, err)
	require.Equal(t, time.August, m)
	require.Equal(t, 3, h.GetDayOfMonth("tria"))
	require.Equal(t, 3, h.GetDayOfMonth("trian"))
	require.Equal(t, 16, h.GetDayOfMonth("deksesa"))
	require.Equal(t, 23, h.GetDayOfMonth("dudek-tria"))
	// no aug invent on helper either
	_, err = h.GetMonth("aug")
	require.Error(t, err)
}

func TestDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewDateCheckFilter()
	// 27 aŭgusto 2014 was Wednesday
	m := rules.NewRuleMatch(nil, nil, 0, 20, "Estis {realDay}, ne {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "aŭgusto", "day": "27", "weekDay": "lundo",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "merkredo")
	require.Contains(t, out.GetMessage(), "lundo")

	// 26 aŭgusto 2014 was Tuesday = mardo
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "aŭgusto", "day": "26", "weekDay": "mardo",
	}, 0, nil, nil)
	require.Nil(t, ok)

	// Numeric month twin: 2022-03-12 Saturday correct / Sunday wrong
	match := rules.NewRuleMatch(nil, nil, 0, 10, "{realDay}/{day}")
	require.Nil(t, f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "sabato",
	}, -1, nil, nil))
	bad := f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "dimanĉo",
	}, -1, nil, nil)
	require.NotNil(t, bad)
	require.Contains(t, bad.GetMessage(), "sabato")
	require.Contains(t, bad.GetMessage(), "dimanĉo")

	// Spelled day-of-month via GetDayOfMonthOptional: 2014-08-27 = "dudek-sepa"
	spelled := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "aŭgusto", "day": "dudek-sepa", "weekDay": "merkredo",
	}, 0, nil, nil)
	require.Nil(t, spelled)
	spelledBad := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "aŭgusto", "day": "dudek-sepa", "weekDay": "lundo",
	}, 0, nil, nil)
	require.NotNil(t, spelledBad)
	require.Contains(t, spelledBad.GetMessage(), "merkredo")
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
	// invent Latin "aug" must also skip (RuntimeException path), not accept as August
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "aug", "day": "27", "weekDay": "lundo",
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

func TestEODateCheckFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.eo.DateCheckFilter"))
}
