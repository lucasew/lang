package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter(t *testing.T) {
	f := NewDateCheckFilter()
	d, err := f.GetDayOfWeekJava("Monday")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	m, err := f.GetMonth("January")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	require.Equal(t, "Friday", f.GetDayOfWeekName(2014, 8, 29))
}

func TestDateCheckFilter_AcceptIncompleteArgs(t *testing.T) {
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 5, "msg")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"year": "2014"}, 0, nil, nil)
	})
}

func TestDateCheckFilter_AcceptWrongWeekday(t *testing.T) {
	// 2014-08-23 is Saturday; claiming Sunday keeps the match
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "wrong {realDay} not {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "Sunday",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "Saturday")
}

func TestDateCheckFilter_AcceptCorrectWeekday(t *testing.T) {
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "Saturday",
	}, 0, nil, nil)
	require.Nil(t, out)
}

func TestDateCheckFilter_DayStrLikeOriginal(t *testing.T) {
	require.Equal(t, "22", DayStrLikeOriginal("22", "22"))
	require.Equal(t, "22nd", DayStrLikeOriginal("22", "22nd"))
	require.Equal(t, "1st", DayStrLikeOriginal("1", "1st"))
	require.Equal(t, "3rd", DayStrLikeOriginal("3", "third"))
	require.Equal(t, "11th", DayStrLikeOriginal("11", "eleventh"))
}

func TestDateCheckFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.en.DateCheckFilter"))
}
