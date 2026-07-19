package de

// Twin of DateCheckFilterTest.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter_GetDayOfWeek1(t *testing.T) {
	f := NewDateCheckFilter()
	// Java Calendar: Sunday=1 … Saturday=7
	d, err := f.GetDayOfWeekJava("So")
	require.NoError(t, err)
	require.Equal(t, 1, d)
	d, err = f.GetDayOfWeekJava("Mo")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("mo")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("Mon.")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("Montag")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	d, err = f.GetDayOfWeekJava("Di")
	require.NoError(t, err)
	require.Equal(t, 3, d)
	d, err = f.GetDayOfWeekJava("Fr")
	require.NoError(t, err)
	require.Equal(t, 6, d)
	d, err = f.GetDayOfWeekJava("Samstag")
	require.NoError(t, err)
	require.Equal(t, 7, d)
	d, err = f.GetDayOfWeekJava("Sonnabend")
	require.NoError(t, err)
	require.Equal(t, 7, d)
}

func TestDateCheckFilter_GetDayOfWeek2(t *testing.T) {
	f := NewDateCheckFilter()
	// 2014-08-29 = Friday, 2014-08-30 = Saturday
	require.Equal(t, "Freitag", f.GetDayOfWeekName(2014, 8, 29))
	require.Equal(t, "Samstag", f.GetDayOfWeekName(2014, 8, 30))
}

func TestDateCheckFilter_GetMonth(t *testing.T) {
	f := NewDateCheckFilter()
	m, err := f.GetMonth("Januar")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("Jan")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("Jan.")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("Dezember")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	m, err = f.GetMonth("Dez")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	m, err = f.GetMonth("DEZEMBER")
	require.NoError(t, err)
	require.Equal(t, 12, m)
}

func TestDateCheckFilter_AcceptIncompleteArgs(t *testing.T) {
	// Java getRequired("weekDay") throws when missing
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 5, "msg")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"year": "2014", "month": "8", "day": "23"}, 0, nil, nil)
	})
}

func TestDateCheckFilter_AcceptWrongWeekday(t *testing.T) {
	// 2014-08-23 is Saturday; claiming Sonntag keeps the match
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "wrong {realDay} not {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "Sonntag",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "Samstag")
}

func TestDateCheckFilter_AcceptCorrectWeekday(t *testing.T) {
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "Samstag",
	}, 0, nil, nil)
	require.Nil(t, out)
}

func TestDateCheckFilter_AdjustSuggestion(t *testing.T) {
	// Java: remove ".," only when index in (5,12); add ".," when short comma form
	require.Equal(t, "Sonntag, den 1.", AdjustDateCheckSuggestion("Sonntag., den 1."))
	require.Equal(t, "So., den 1.", AdjustDateCheckSuggestion("So, den 1."))
	// short "So.," — ".," at index 2, not in (5,12) → unchanged
	require.Equal(t, "So., den 1.", AdjustDateCheckSuggestion("So., den 1."))
}
