package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDMYDateCheckFilter_PrepareArgs(t *testing.T) {
	f := NewDMYDateCheckFilter()
	_, err := f.PrepareArgs(map[string]string{"year": "2020"})
	require.Error(t, err)
	out, err := f.PrepareArgs(map[string]string{"date": "15-01-2020", "weekDay": "mercredi"})
	require.NoError(t, err)
	require.Equal(t, "15", out["day"])
	require.Equal(t, "01", out["month"])
	require.Equal(t, "2020", out["year"])
}

func TestDMYDateCheckFilter_AcceptWrongWeekday(t *testing.T) {
	// 2014-08-23 is Saturday (samedi); date as dd-mm-yyyy
	f := NewDMYDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "wrong {realDay} not {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"date":    "23-08-2014",
		"weekDay": "dimanche",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "samedi")
}

func TestDMYDateCheckFilter_AcceptCorrectWeekday(t *testing.T) {
	f := NewDMYDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"date":    "23-08-2014",
		"weekDay": "samedi",
	}, 0, nil, nil)
	require.Nil(t, out)
}

func TestDMYDateCheckFilter_RejectsYearKey(t *testing.T) {
	f := NewDMYDateCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 1, "msg")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "23-08-2014", "year": "2014", "weekDay": "samedi"}, 0, nil, nil)
	})
}

func TestDMYDateCheckFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.fr.DMYDateCheckFilter"))
}
