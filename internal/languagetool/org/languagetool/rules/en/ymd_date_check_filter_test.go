package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestYMDDateCheckFilter_PrepareArgs(t *testing.T) {
	f := NewYMDDateCheckFilter()
	_, err := f.PrepareArgs(map[string]string{"month": "1"})
	require.Error(t, err)
	out, err := f.PrepareArgs(map[string]string{"date": "1999-12-31", "weekDay": "6"})
	require.NoError(t, err)
	require.Equal(t, "1999", out["year"])
	require.Equal(t, "12", out["month"])
	require.Equal(t, "31", out["day"])
}

func TestYMDDateCheckFilter_AcceptRuleMatch_WrongWeekday(t *testing.T) {
	// 2014-08-23 is Saturday; claiming Sunday keeps the match
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

func TestYMDDateCheckFilter_RejectsYearMonthDayKeys(t *testing.T) {
	f := NewYMDDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 1, "msg")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "2014-08-23", "year": "2014", "weekDay": "Saturday"}, 0, nil, nil)
	})
}

func TestYMDDateCheckFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.en.YMDDateCheckFilter"))
}
