package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Expanded outcome matrix for DateCheckFilter.java day rules
// (equals vs startsWith; no trimSpecialCharacters).
func TestDateCheckFilter_GetDayOfWeek_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	// Java Calendar Sunday=1 … Saturday=7
	cases := []struct {
		in   string
		want int
	}{
		// Monday: startsWith("пн") || equals("понедельник")
		{"пн", 2},
		{"пн.", 2},
		{"Пн", 2},
		{"понедельник", 2},
		{"Понедельник", 2},
		// Tuesday: startsWith("вт")
		{"вт", 3},
		{"вторник", 3},
		// Wednesday: startsWith("ср")
		{"ср", 4},
		{"среда", 4},
		// Thursday: startsWith("чт") || equals("четверг")
		{"чт", 5},
		{"четверг", 5},
		// Friday: equals("пт") || startsWith("пятниц")
		{"пт", 6},
		{"пятница", 6},
		{"пятниц", 6},
		// Saturday: startsWith("сб") || startsWith("суббот")
		{"сб", 7},
		{"суббота", 7},
		{"суббот", 7},
		// Sunday: startsWith("вс") || equals("воскресенье")
		{"вс", 1},
		{"воскресенье", 1},
	}
	for _, tc := range cases {
		d, err := f.GetDayOfWeekJava(tc.in)
		require.NoError(t, err, "day %q", tc.in)
		require.Equal(t, tc.want, d, "day %q", tc.in)
	}
}

// DateCheckFilter Friday uses equals("пт") not startsWith — "пт." must fail.
// Helper uses startsWith("пт") so this is a DateCheckFilter-unique outcome.
func TestDateCheckFilter_GetDayOfWeek_EqualsVsStartsWith(t *testing.T) {
	f := NewDateCheckFilter()
	// equals("пт") only — trailing dot not accepted
	_, err := f.GetDayOfWeekJava("пт.")
	require.Error(t, err, "пт. must fail (equals пт only, not startsWith)")
	// Monday uses startsWith("пн") so trailing punct works
	d, err := f.GetDayOfWeekJava("пн.")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	// "пон" alone is not equals("понедельник") and not startsWith("пн")
	_, err = f.GetDayOfWeekJava("пон")
	require.Error(t, err, "пон alone is not accepted by DateCheckFilter")
	// "чет" alone is not equals("четверг") and not startsWith("чт")
	_, err = f.GetDayOfWeekJava("чет")
	require.Error(t, err, "чет alone is not accepted by DateCheckFilter")
	// "вос" alone is not equals("воскресенье") and not startsWith("вс")
	_, err = f.GetDayOfWeekJava("вос")
	require.Error(t, err, "вос alone is not accepted by DateCheckFilter")
}

func TestDateCheckFilter_GetDayOfWeek_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	for _, in := range []string{"", "xyz", "monday", "Mon"} {
		_, err := f.GetDayOfWeekJava(in)
		require.Error(t, err, "expected error for day %q", in)
	}
}

// Expanded month matrix: nominative/genitive/abbrev equals + Roman I..XII
// (Romans are real Java DateCheckFilter branches, case-sensitive).
func TestDateCheckFilter_GetMonth_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		// January
		{"январь", 1}, {"I", 1}, {"января", 1}, {"янв", 1}, {"Январь", 1},
		// February
		{"февраль", 2}, {"II", 2}, {"февраля", 2}, {"фев", 2},
		// March
		{"март", 3}, {"III", 3}, {"марта", 3}, {"мар", 3},
		// April
		{"апрель", 4}, {"IV", 4}, {"апреля", 4}, {"апр", 4},
		// May (no short "май" genitive beyond мая; no abbr beyond nominative)
		{"май", 5}, {"V", 5}, {"мая", 5},
		// June — includes DateCheckFilter-unique "ин"
		{"июнь", 6}, {"VI", 6}, {"июня", 6}, {"ин", 6},
		// July — includes DateCheckFilter-unique "ил"
		{"июль", 7}, {"VII", 7}, {"июля", 7}, {"ил", 7},
		// August
		{"август", 8}, {"VIII", 8}, {"августа", 8}, {"авг", 8},
		// September
		{"сентябрь", 9}, {"IX", 9}, {"сентября", 9}, {"сен", 9},
		// October
		{"октябрь", 10}, {"X", 10}, {"октября", 10}, {"окт", 10},
		// November
		{"ноябрь", 11}, {"XI", 11}, {"ноября", 11}, {"ноя", 11},
		// December
		{"декабрь", 12}, {"XII", 12}, {"декабря", 12}, {"дек", 12},
		{"Декабрь", 12}, {"ДЕКАБРЬ", 12},
	}
	for _, tc := range cases {
		m, err := f.GetMonth(tc.in)
		require.NoError(t, err, "month %q", tc.in)
		require.Equal(t, tc.want, m, "month %q", tc.in)
	}
}

// Roman numerals: case-sensitive on original monthStr (not lowercased).
func TestDateCheckFilter_GetMonth_RomanCaseSensitive(t *testing.T) {
	f := NewDateCheckFilter()
	// uppercase Romans accepted
	for _, in := range []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"} {
		_, err := f.GetMonth(in)
		require.NoError(t, err, "Roman %q", in)
	}
	// lowercase / mixed must fail (monthStr.equals("I") etc.)
	for _, in := range []string{"i", "ii", "xii", "Xii", "Viii"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "lowercase/mixed Roman %q must fail", in)
	}
}

// DateCheckFilter uses exact equals for Russian month names (after toLowerCase),
// not startsWith — so longer forms that DateFilterHelper accepts may fail here.
func TestDateCheckFilter_GetMonth_EqualsNotPrefix(t *testing.T) {
	f := NewDateCheckFilter()
	// prefix-only forms that are NOT full equals in DateCheckFilter
	for _, in := range []string{"янва", "февра", "апрел", "июн", "июл", "авгус", "сентяб", "октяб", "нояб", "декаб"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "partial form %q must fail equals-only DateCheckFilter", in)
	}
	// accepted abbrevs that DateCheckFilter lists
	for _, in := range []string{"янв", "фев", "мар", "апр", "авг", "сен", "окт", "ноя", "дек", "ин", "ил"} {
		_, err := f.GetMonth(in)
		require.NoError(t, err, "abbrev %q", in)
	}
}

func TestDateCheckFilter_GetMonth_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	for _, in := range []string{"", "xyz", "february", "May", "13"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "expected error for month %q", in)
	}
}

func TestDateCheckFilter_Helpers(t *testing.T) {
	f := NewDateCheckFilter()
	m, err := f.GetMonth("августа")
	require.NoError(t, err)
	require.Equal(t, 8, m)
	// Roman VIII is real DateCheckFilter.java (case-sensitive equals)
	m, err = f.GetMonth("VIII")
	require.NoError(t, err)
	require.Equal(t, 8, m)
	// 27 августа 2014 was Wednesday
	require.Equal(t, "среда", f.GetDayOfWeekName(2014, 8, 27))
	jd, err := f.GetDayOfWeekJava("вторник")
	require.NoError(t, err)
	require.Equal(t, 3, jd)
}

func TestDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewDateCheckFilter()
	// 27 августа 2014 was Wednesday, not вторник
	m := rules.NewRuleMatch(nil, nil, 0, 20, "День {realDay}, не {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "августа", "day": "27", "weekDay": "вторник",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "среда")
	require.Contains(t, out.GetMessage(), "вторник")

	// 26 августа 2014 was Tuesday
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "августа", "day": "26", "weekDay": "вторник",
	}, 0, nil, nil)
	require.Nil(t, ok)

	// Roman month path through AcceptRuleMatch (Abstract trims then getMonth)
	// 1 I 2014 was Wednesday
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "I", "day": "1", "weekDay": "вторник",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "среда")
}

func TestRUDateCheckFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.ru.DateCheckFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
}
