package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter(t *testing.T) {
	f := NewDateCheckFilter()
	d, err := f.GetDayOfWeekJava("lunes")
	require.NoError(t, err)
	require.Equal(t, 2, d)
	m, err := f.GetMonth("enero")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	require.Equal(t, "viernes", f.GetDayOfWeekName(2014, 8, 29))
	require.Equal(t, "sábado", f.GetDayOfWeekName(2014, 8, 23))
}

func TestDateCheckFilter_AcceptWrongWeekday(t *testing.T) {
	// 2014-08-23 is Saturday (sábado); claiming domingo keeps the match
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "wrong {realDay} not {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "domingo",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "sábado")
}

func TestDateCheckFilter_AcceptCorrectWeekday(t *testing.T) {
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "sábado",
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
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.es.DateCheckFilter"))
}
