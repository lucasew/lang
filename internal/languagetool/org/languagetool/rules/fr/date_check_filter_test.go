package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter(t *testing.T) {
	f := NewDateCheckFilter()
	d, err := f.GetDayOfWeekJava("lundi")
	require.NoError(t, err)
	require.Equal(t, 2, d) // Java Monday=2
	m, err := f.GetMonth("janvier")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	require.Equal(t, "vendredi", f.GetDayOfWeekName(2014, 8, 29))
	require.Equal(t, "samedi", f.GetDayOfWeekName(2014, 8, 23))
}

func TestDateCheckFilter_AcceptWrongWeekday(t *testing.T) {
	// 2014-08-23 is Saturday (samedi); claiming dimanche keeps the match
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "wrong {realDay} not {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "dimanche",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "samedi")
}

func TestDateCheckFilter_AcceptCorrectWeekday(t *testing.T) {
	f := NewDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "8", "day": "23", "weekDay": "samedi",
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
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.fr.DateCheckFilter"))
}
