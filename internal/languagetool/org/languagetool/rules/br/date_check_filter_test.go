package br

import (
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Expanded outcome matrix for DateCheckFilter.getDayOfWeek(String).
// Java Calendar Sunday=1 … Saturday=7.
// endsWith sul/lun/meurzh/merc’her; equals yaou/diriaou; endsWith gwener/sadorn.
func TestDateCheckFilter_GetDayOfWeek_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		// Sunday: endsWith("sul")
		{"sul", 1},
		{"Sul", 1},
		{"Disul", 1},
		// Monday: endsWith("lun")
		{"lun", 2},
		{"Lun", 2},
		{"Dilun", 2},
		// Tuesday: endsWith("meurzh")
		{"meurzh", 3},
		{"Meurzh", 3},
		{"Dimeurzh", 3},
		// Wednesday: endsWith("merc’her") U+2019
		{"merc’her", 4},
		{"Merc’her", 4},
		{"Dimerc’her", 4},
		// Thursday: equals yaou / diriaou only
		{"yaou", 5},
		{"Yaou", 5},
		{"diriaou", 5},
		{"Diriaou", 5},
		// Friday: endsWith("gwener")
		{"gwener", 6},
		{"Gwener", 6},
		{"Digwener", 6},
		// Saturday: endsWith("sadorn")
		{"sadorn", 7},
		{"Sadorn", 7},
		{"Disadorn", 7},
	}
	for _, tc := range cases {
		d, err := f.GetDayOfWeekJava(tc.in)
		require.NoError(t, err, "day %q", tc.in)
		require.Equal(t, tc.want, d, "day %q", tc.in)
	}
}

func TestDateCheckFilter_GetDayOfWeek_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	// English / unknown / ASCII-apostrophe invent must fail.
	for _, in := range []string{"", "xyz", "monday", "Mon", "thursday", "Sunday", "merc'her"} {
		_, err := f.GetDayOfWeekJava(in)
		require.Error(t, err, "expected error for day %q", in)
		require.Contains(t, err.Error(), "could not find day of week for '"+in+"'")
	}
	// "yaou" is equals-only: prefix/suffix invents fail
	for _, in := range []string{"yaoua", "xyaou"} {
		_, err := f.GetDayOfWeekJava(in)
		require.Error(t, err, "expected error for day %q", in)
	}
}

// getDayOfWeek(Calendar): Locale.UK LONG English → Breton names.
func TestDateCheckFilter_GetDayOfWeek_CalendarFormat(t *testing.T) {
	f := NewDateCheckFilter()
	// 2022-03-20..26: Sunday … Saturday
	require.Equal(t, "Sul", f.GetDayOfWeekName(2022, 3, 20))
	require.Equal(t, "Lun", f.GetDayOfWeekName(2022, 3, 21))
	require.Equal(t, "Meurzh", f.GetDayOfWeekName(2022, 3, 22))
	require.Equal(t, "Merc’her", f.GetDayOfWeekName(2022, 3, 23))
	require.Equal(t, "Yaou", f.GetDayOfWeekName(2022, 3, 24))
	require.Equal(t, "Gwener", f.GetDayOfWeekName(2022, 3, 25))
	require.Equal(t, "Sadorn", f.GetDayOfWeekName(2022, 3, 26))
	// Known dates used in AcceptRuleMatch twins
	require.Equal(t, "Merc’her", f.GetDayOfWeekName(2014, 8, 27))
	require.Equal(t, "Meurzh", f.GetDayOfWeekName(2014, 8, 26))
	require.Equal(t, "Lun", f.GetDayOfWeekName(2024, 1, 1))
}

// Full Breton month equals map.
func TestDateCheckFilter_GetMonth_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		{"genver", 1}, {"Genver", 1},
		{"c’hwevrer", 2}, {"C’hwevrer", 2},
		{"meurzh", 3}, {"Meurzh", 3},
		{"ebrel", 4},
		{"mae", 5},
		{"mezheven", 6}, {"even", 6}, {"Even", 6},
		{"gouere", 7}, {"gouhere", 7}, {"Gouere", 7},
		{"eost", 8}, {"Eost", 8},
		{"gwengolo", 9},
		{"here", 10},
		{"du", 11},
		{"kerzu", 12}, {"Kerzu", 12},
	}
	for _, tc := range cases {
		got, err := f.GetMonth(tc.in)
		require.NoError(t, err, "month %q", tc.in)
		require.Equal(t, tc.want, got, "month %q", tc.in)
	}
}

func TestDateCheckFilter_GetMonth_NoASCIIApostropheInvent(t *testing.T) {
	// Java only mon.equals("c’hwevrer") with U+2019 — ASCII ' must fail.
	f := NewDateCheckFilter()
	for _, in := range []string{"c'hwevrer", "C'hwevrer"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "ASCII apostrophe invent %q must fail", in)
		require.Contains(t, err.Error(), "could not find month '"+in+"'")
	}
}

func TestDateCheckFilter_GetMonth_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	for _, in := range []string{"", "xyz", "january", "august", "aug", "may", "June", "december", "gen"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "expected error for month %q", in)
		require.Contains(t, err.Error(), "could not find month '"+in+"'")
	}
}

// Day-of-month: soft mutation t→d / p→b; strip "vet"; full map; 0 unknown.
func TestDateCheckFilter_GetDayOfMonth_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		// base map (U+2019 forms)
		{"c’hentañ", 1},
		{"unan", 1},
		{"daou", 2},
		{"eil", 2},
		{"dri", 3},
		{"drede", 3},
		{"deir", 3},
		{"bevar", 4},
		{"bemp", 5},
		{"bem", 5},
		{"c’hwerc’h", 6},
		{"seizh", 7},
		{"eizh", 8},
		{"nav", 9},
		{"na", 9},
		{"dek", 10},
		{"unnek", 11},
		{"daouzek", 12},
		{"drizek", 13},
		{"bevarzek", 14},
		{"bemzek", 15},
		{"c’hwezek", 16},
		{"seitek", 17},
		{"driwec’h", 18},
		{"naontek", 19},
		{"ugent", 20},
		{"dregont", 30},
		// case
		{"Unan", 1},
		{"DEK", 10},
		// trailing "vet"
		{"bempvet", 5},
		{"bemvet", 5},
		{"navet", 9},
		{"daouvet", 2},
		{"c’hentañvet", 1},
		// soft mutation t→d
		{"taou", 2}, // → daou
		{"tri", 3},  // → dri
		{"trede", 3},
		{"teir", 3},
		{"trizek", 13}, // → drizek
		{"triwec’h", 18},
		{"tregont", 30},
		// soft mutation p→b
		{"pevar", 4}, // → bevar
		{"pemp", 5},
		{"pem", 5},
		{"pevarzek", 14},
		{"pemzek", 15},
		// mutation + vet
		{"taouvet", 2},
		{"pempvet", 5},
		// unknown
		{"xyz", 0},
		{"tridek", 0},
		// ASCII apostrophe invent must not match
		{"c'hentañ", 0},
		{"c'hwerc'h", 0},
		{"c'hwezek", 0},
		{"driwec'h", 0},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, f.GetDayOfMonth(tc.in), "dayOfMonth %q", tc.in)
	}
}

// Helper surface still matches filter for localization primitives.
func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	wd, err := h.GetDayOfWeek("Dilun")
	require.NoError(t, err)
	require.Equal(t, time.Monday, wd)
	wd, err = h.GetDayOfWeek("diriaou")
	require.NoError(t, err)
	require.Equal(t, time.Thursday, wd)
	m, err := h.GetMonth("eost")
	require.NoError(t, err)
	require.Equal(t, time.August, m)
	m, err = h.GetMonth("even")
	require.NoError(t, err)
	require.Equal(t, time.June, m)
	require.Equal(t, 1, h.GetDayOfMonth("unan"))
	require.Equal(t, 5, h.GetDayOfMonth("bempvet"))
	require.Equal(t, 2, h.GetDayOfMonth("taou"))
	// no ASCII invent on helper either
	_, err = h.GetMonth("c'hwevrer")
	require.Error(t, err)
	_, err = h.GetDayOfWeek("merc'her")
	require.Error(t, err)
}

func TestDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewDateCheckFilter()
	// 27 eost 2014 was Wednesday; claim Lun (Monday) → mismatch
	m := rules.NewRuleMatch(nil, nil, 0, 20, "Said {day} was {realDay}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "27", "weekDay": "Lun",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "Merc’her")
	require.Contains(t, out.GetMessage(), "Lun")

	// 26 eost 2014 was Tuesday = Meurzh
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "26", "weekDay": "Meurzh",
	}, 0, nil, nil)
	require.Nil(t, ok)

	// Numeric month twin: 2022-03-12 Saturday correct / Sunday wrong
	match := rules.NewRuleMatch(nil, nil, 0, 10, "{realDay}/{day}")
	require.Nil(t, f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "Sadorn",
	}, -1, nil, nil))
	bad := f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "Sul",
	}, -1, nil, nil)
	require.NotNil(t, bad)
	require.Contains(t, bad.GetMessage(), "Sadorn")
	require.Contains(t, bad.GetMessage(), "Sul")

	// Spelled day-of-month via GetDayOfMonthOptional: 2014-08-27 — use numeric 27
	// (Breton map has no compound 27; ugent=20, seizh=7 only).
	// endsWith weekday form: "Dilun" vs real Merc’her
	prefixBad := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "27", "weekDay": "Dilun",
	}, 0, nil, nil)
	require.NotNil(t, prefixBad)
	require.Contains(t, prefixBad.GetMessage(), "Merc’her")
	require.Contains(t, prefixBad.GetMessage(), "Lun") // {day} is FormatDayOfWeek of claimed weekday

	// Spelled day matching: unan = 1 → 2014-08-01 was Friday = Gwener
	spelled := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "unan", "weekDay": "Gwener",
	}, 0, nil, nil)
	require.Nil(t, spelled)
	spelledBad := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "unan", "weekDay": "Lun",
	}, 0, nil, nil)
	require.NotNil(t, spelledBad)
	require.Contains(t, spelledBad.GetMessage(), "Gwener")

	// soft-mutated spelled day: taou → 2 → 2014-08-02 Saturday
	mut := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "taou", "weekDay": "Sadorn",
	}, 0, nil, nil)
	require.Nil(t, mut)

	// diriaou / yaou Thursday: 2014-08-28 was Thursday
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "28", "weekDay": "diriaou",
	}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "28", "weekDay": "yaou",
	}, 0, nil, nil))
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
		"year": "2022", "month": "notamonth", "day": "12", "weekDay": "Sadorn",
	}, -1, nil, nil)
	require.Nil(t, out)
	// English "august" must skip (RuntimeException path), not invent August
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "august", "day": "27", "weekDay": "Lun",
	}, -1, nil, nil)
	require.Nil(t, out)
	// ASCII invent for February must skip
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "c'hwevrer", "day": "1", "weekDay": "Sadorn",
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

func TestBRDateCheckFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.br.DateCheckFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
}
