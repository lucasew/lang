package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter_Helpers(t *testing.T) {
	f := NewDateCheckFilter()
	m, err := f.GetMonth("sierpnia")
	require.NoError(t, err)
	require.Equal(t, 8, m)
	// 27 sierpnia 2014 was Wednesday
	require.Equal(t, "środa", f.GetDayOfWeekName(2014, 8, 27))
	jd, err := f.GetDayOfWeekJava("wtorek")
	require.NoError(t, err)
	require.Equal(t, 3, jd) // Java Calendar: Sunday=1, Tuesday=3
}

func TestDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewDateCheckFilter()
	// Grammar example: "wtorek, 27 sierpnia 2014" — 27 Aug 2014 was Wednesday, not Tuesday
	m := rules.NewRuleMatch(nil, nil, 0, 20, "Dzień {realDay}, nie {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "sierpnia", "day": "27", "weekDay": "wtorek",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "środa")
	require.Contains(t, out.GetMessage(), "wtorek")

	// Correct: 26 sierpnia 2014 was Tuesday
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "sierpnia", "day": "26", "weekDay": "wtorek",
	}, 0, nil, nil)
	require.Nil(t, ok)

	// Numeric month form (second grammar rule)
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "08", "day": "27", "weekDay": "wtorek",
	}, 0, nil, nil)
	require.NotNil(t, out)
}

func TestPLDateCheckFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.pl.DateCheckFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg {realDay}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "sierpnia", "day": "27", "weekDay": "niedziela",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "środa")
}
