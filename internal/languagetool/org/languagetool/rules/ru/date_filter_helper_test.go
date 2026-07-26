package ru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateFilterHelper_GetDayOfWeek(t *testing.T) {
	h := NewDateFilterHelper()
	cases := []struct {
		in   string
		want time.Weekday
	}{
		// Java startsWith prefixes in order
		{"суб", time.Saturday},
		{"суббота", time.Saturday},
		{"вс", time.Sunday},
		{"вс.", time.Sunday},
		{"вос", time.Sunday},
		{"воскресенье", time.Sunday},
		{"пн", time.Monday},
		{"пн.", time.Monday},
		{"пон", time.Monday},
		{"понедельник", time.Monday},
		{"вт", time.Tuesday},
		{"вторник", time.Tuesday},
		{"ср", time.Wednesday},
		{"среда", time.Wednesday},
		{"чт", time.Thursday},
		{"чет", time.Thursday},
		{"четверг", time.Thursday},
		{"пт", time.Friday},
		{"пт.", time.Friday},
		{"пят", time.Friday},
		{"пятница", time.Friday},
		{"сб", time.Saturday},
		{"сб.", time.Saturday},
		// mixed case after toLowerCase
		{"Понедельник", time.Monday},
		{"СУББОТА", time.Saturday},
		{"Вс", time.Sunday},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := h.GetDayOfWeek(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestDateFilterHelper_GetDayOfWeek_Unknown(t *testing.T) {
	h := NewDateFilterHelper()
	for _, in := range []string{"xyz", "", "monday", "Saturday"} {
		_, err := h.GetDayOfWeek(in)
		require.Error(t, err, "expected error for %q", in)
	}
}

func TestDateFilterHelper_GetDayOfWeek_TrimSpecialCharacters(t *testing.T) {
	h := NewDateFilterHelper()
	// soft hyphen U+00AD stripped by StringTools.trimSpecialCharacters
	d, err := h.GetDayOfWeek("пон\u00ADедельник")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	d, err = h.GetDayOfWeek("суб\u00ADбота")
	require.NoError(t, err)
	require.Equal(t, time.Saturday, d)
	d, err = h.GetDayOfWeek("вос\u00ADкресенье")
	require.NoError(t, err)
	require.Equal(t, time.Sunday, d)
}

func TestDateFilterHelper_GetDayOfWeek_SubBeforeSb(t *testing.T) {
	// Java checks "суб" before "сб" — both Saturday
	h := NewDateFilterHelper()
	d, err := h.GetDayOfWeek("суббота")
	require.NoError(t, err)
	require.Equal(t, time.Saturday, d)
	d, err = h.GetDayOfWeek("сб")
	require.NoError(t, err)
	require.Equal(t, time.Saturday, d)
}

func TestDateFilterHelper_GetMonth(t *testing.T) {
	h := NewDateFilterHelper()
	cases := []struct {
		in   string
		want time.Month
	}{
		{"янв", time.January},
		{"январь", time.January},
		{"января", time.January},
		{"фев", time.February},
		{"февраль", time.February},
		{"февраля", time.February},
		{"мар", time.March},
		{"март", time.March},
		{"марта", time.March},
		{"апр", time.April},
		{"апрель", time.April},
		{"апреля", time.April},
		{"май", time.May},
		{"мая", time.May},
		{"июн", time.June},
		{"июнь", time.June},
		{"июня", time.June},
		{"июл", time.July},
		{"июль", time.July},
		{"июля", time.July},
		{"авг", time.August},
		{"август", time.August},
		{"августа", time.August},
		{"сен", time.September},
		{"сентябрь", time.September},
		{"сентября", time.September},
		{"окт", time.October},
		{"октябрь", time.October},
		{"октября", time.October},
		{"ноя", time.November},
		{"ноябрь", time.November},
		{"ноября", time.November},
		{"дек", time.December},
		{"декабрь", time.December},
		{"декабря", time.December},
		{"ФЕВ", time.February},
		{"Декабрь", time.December},
		{"ДЕКАБРЬ", time.December},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := h.GetMonth(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestDateFilterHelper_GetMonth_Unknown(t *testing.T) {
	h := NewDateFilterHelper()
	// Unknown / English invent
	for _, in := range []string{"xyz", "", "february", "May"} {
		_, err := h.GetMonth(in)
		require.Error(t, err, "expected error for %q", in)
	}
	// Roman numerals invent (only on DateCheckFilter.java, not DateFilterHelper)
	for _, in := range []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"} {
		_, err := h.GetMonth(in)
		require.Error(t, err, "expected error for Roman %q (not in DateFilterHelper)", in)
	}
	// invent short forms from old DateCheckFilter equals branches
	for _, in := range []string{"ин", "ил"} {
		_, err := h.GetMonth(in)
		require.Error(t, err, "expected error for invent short %q", in)
	}
}

func TestDateFilterHelper_GetMonth_TrimSpecialCharacters(t *testing.T) {
	h := NewDateFilterHelper()
	m, err := h.GetMonth("фев\u00ADраль")
	require.NoError(t, err)
	require.Equal(t, time.February, m)
	m, err = h.GetMonth("авг\u00ADуста")
	require.NoError(t, err)
	require.Equal(t, time.August, m)
	m, err = h.GetMonth("сен\u00ADтябрь")
	require.NoError(t, err)
	require.Equal(t, time.September, m)
}

func TestDateFilterHelper_GetDayOfWeekName(t *testing.T) {
	h := NewDateFilterHelper()
	// Fixed UTC dates with known weekdays
	cases := []struct {
		t    time.Time
		want string
	}{
		{time.Date(2014, 8, 24, 0, 0, 0, 0, time.UTC), "воскресенье"}, // Sunday
		{time.Date(2014, 8, 25, 0, 0, 0, 0, time.UTC), "понедельник"}, // Monday
		{time.Date(2014, 8, 26, 0, 0, 0, 0, time.UTC), "вторник"},     // Tuesday
		{time.Date(2014, 8, 27, 0, 0, 0, 0, time.UTC), "среда"},       // Wednesday
		{time.Date(2014, 8, 28, 0, 0, 0, 0, time.UTC), "четверг"},     // Thursday
		{time.Date(2014, 8, 29, 0, 0, 0, 0, time.UTC), "пятница"},     // Friday
		{time.Date(2014, 8, 23, 0, 0, 0, 0, time.UTC), "суббота"},     // Saturday
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			require.Equal(t, tc.want, h.GetDayOfWeekName(tc.t))
		})
	}
}

func TestDateFilterHelper_GetCalendar(t *testing.T) {
	h := NewDateFilterHelper()
	before := time.Now().Add(-time.Second)
	got := h.GetCalendar()
	after := time.Now().Add(time.Second)
	require.False(t, got.Before(before))
	require.False(t, got.After(after))
}

// Smoke kept for continuity with earlier partial twin tests.
func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	d, err := h.GetDayOfWeek("понедельник")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	m, err := h.GetMonth("января")
	require.NoError(t, err)
	require.Equal(t, time.January, m)
}
