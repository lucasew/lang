package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Expanded outcome matrix for DateCheckFilter.java day rules
// (startsWith prefixes + full equals; toLowerCase).
func TestDateCheckFilter_GetDayOfWeek_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	// Java Calendar Sunday=1 … Saturday=7
	cases := []struct {
		in   string
		want int
	}{
		// Monday: startsWith("по") || equals("понедељак")
		{"по", 2},
		{"понедељак", 2},
		{"Понедељак", 2},
		{"ПОНЕДЕЉАК", 2},
		{"пон", 2},
		// Tuesday: startsWith("ут")
		{"ут", 3},
		{"уторак", 3},
		{"Уторак", 3},
		// Wednesday: startsWith("ср")
		{"ср", 4},
		{"среда", 4},
		// Thursday: startsWith("че") || equals("четвртак")
		{"че", 5},
		{"чет", 5},
		{"четвртак", 5},
		{"Четвртак", 5},
		// Friday: startsWith("пе") || equals("петак")
		{"пе", 6},
		{"петак", 6},
		{"Петак", 6},
		// Saturday: startsWith("су") || equals("субота")
		{"су", 7},
		{"субота", 7},
		{"Субота", 7},
		// Sunday: startsWith("не") || equals("недеља")
		{"не", 1},
		{"недеља", 1},
		{"Недеља", 1},
	}
	for _, tc := range cases {
		d, err := f.GetDayOfWeekJava(tc.in)
		require.NoError(t, err, "day %q", tc.in)
		require.Equal(t, tc.want, d, "day %q", tc.in)
	}
}

func TestDateCheckFilter_GetDayOfWeek_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	for _, in := range []string{"", "xyz", "monday", "Mon"} {
		_, err := f.GetDayOfWeekJava(in)
		require.Error(t, err, "expected error for day %q", in)
	}
	// Exact Java RuntimeException message
	_, err := f.GetDayOfWeekJava("xyz")
	require.EqualError(t, err, "Редни број дана у недељи за 'xyz' не постоји.")
}

// startsWith("по") accepts forms with trailing punctuation.
func TestDateCheckFilter_GetDayOfWeek_PrefixPunct(t *testing.T) {
	f := NewDateCheckFilter()
	d, err := f.GetDayOfWeekJava("пон.")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("ут.")
	require.NoError(t, err)
	require.Equal(t, 3, d)
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
		{"јануар", 1}, {"I", 1}, {"јануара", 1}, {"јан", 1}, {"Јануар", 1},
		// February
		{"фебруар", 2}, {"II", 2}, {"фебруара", 2}, {"феб", 2},
		// March
		{"март", 3}, {"III", 3}, {"марта", 3}, {"мар", 3},
		// April
		{"април", 4}, {"IV", 4}, {"априла", 4}, {"апр", 4},
		// May (no short abbrev beyond nominative/genitive)
		{"мај", 5}, {"V", 5}, {"маја", 5},
		// June
		{"јун", 6}, {"VI", 6}, {"јуна", 6}, {"Јун", 6},
		// July
		{"јул", 7}, {"VII", 7}, {"јула", 7},
		// August
		{"август", 8}, {"VIII", 8}, {"августа", 8}, {"авг", 8},
		// September
		{"септембар", 9}, {"IX", 9}, {"септембра", 9}, {"сеп", 9},
		// October
		{"октобар", 10}, {"X", 10}, {"октобра", 10}, {"окт", 10},
		// November
		{"новембар", 11}, {"XI", 11}, {"новембра", 11}, {"нов", 11},
		// December
		{"децембар", 12}, {"XII", 12}, {"децембра", 12}, {"дец", 12},
		{"Децембар", 12}, {"ДЕЦЕМБАР", 12},
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
	for _, in := range []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"} {
		_, err := f.GetMonth(in)
		require.NoError(t, err, "Roman %q", in)
	}
	for _, in := range []string{"i", "ii", "xii", "Xii", "Viii"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "lowercase/mixed Roman %q must fail", in)
	}
}

// DateCheckFilter uses exact equals for Serbian month names (after toLowerCase),
// not startsWith.
func TestDateCheckFilter_GetMonth_EqualsNotPrefix(t *testing.T) {
	f := NewDateCheckFilter()
	for _, in := range []string{"јану", "фебру", "април.", "јуни", "јули", "авгус", "септем", "октоб", "новем", "децем"} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "partial form %q must fail equals-only DateCheckFilter", in)
	}
	// accepted abbrevs that DateCheckFilter lists
	for _, in := range []string{"јан", "феб", "мар", "апр", "авг", "сеп", "окт", "нов", "дец"} {
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
	_, err := f.GetMonth("xyz")
	require.EqualError(t, err, "Месец 'xyz' не постоји.")
}

func TestDateCheckFilter_Helpers(t *testing.T) {
	f := NewDateCheckFilter()
	// Java Calendar Sunday=1
	d, err := f.GetDayOfWeekJava("недеља")
	require.NoError(t, err)
	require.Equal(t, 1, d)
	d, err = f.GetDayOfWeekJava("по")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	m, err := f.GetMonth("јануар")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("XII")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	// 2014-08-29 was Friday; 2014-08-23 was Saturday (Locale "sr" LONG)
	require.Equal(t, "петак", f.GetDayOfWeekName(2014, 8, 29))
	require.Equal(t, "субота", f.GetDayOfWeekName(2014, 8, 23))
}

func TestDateCheckFilter_GetDayOfWeekName_AllDays(t *testing.T) {
	f := NewDateCheckFilter()
	// 2014-08-24..30 is Sun..Sat
	want := []string{"недеља", "понедељак", "уторак", "среда", "четвртак", "петак", "субота"}
	for i, name := range want {
		day := 24 + i // Aug 24 2014 = Sunday
		require.Equal(t, name, f.GetDayOfWeekName(2014, 8, day), "day offset %d", i)
	}
}

func TestDateCheckFilter_AcceptWrongWeekday(t *testing.T) {
	// 2014-08-23 is Saturday (субота); claiming недеља keeps the match
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "wrong {realDay} not {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "недеља",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "субота")
	require.Contains(t, out.GetMessage(), "недеља")
}

func TestDateCheckFilter_AcceptCorrectWeekday(t *testing.T) {
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "субота",
	}, 0, nil, nil)
	require.Nil(t, out)
}

func TestDateCheckFilter_AcceptRuleMatch_NamedMonth(t *testing.T) {
	f := NewDateCheckFilter()
	// 27. августа 2014 was Wednesday (среда), not уторак
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 20, "Дан {realDay}, не {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "августа", "day": "27", "weekDay": "уторак",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "среда")
	require.Contains(t, out.GetMessage(), "уторак")

	// 26. августа 2014 was Tuesday
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "августа", "day": "26", "weekDay": "уторак",
	}, 0, nil, nil)
	require.Nil(t, ok)

	// Roman month path through AcceptRuleMatch (Abstract trims then getMonth)
	// 1 I 2014 was Wednesday
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "I", "day": "1", "weekDay": "уторак",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "среда")
}

func TestDateCheckFilter_AcceptRuleMatch_UnknownWeekdaySkips(t *testing.T) {
	// Java RuntimeException from getDayOfWeek → logger.warn + return null
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "notaday",
	}, 0, nil, nil)
	require.Nil(t, out)
}

func TestDateCheckFilter_MissingWeekDayPanics(t *testing.T) {
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 5, "msg")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"year": "2014"}, 0, nil, nil)
	})
}

func TestDateCheckFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.sr.DateCheckFilter"))
}
