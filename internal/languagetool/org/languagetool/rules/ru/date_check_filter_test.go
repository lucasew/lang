package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter_Helpers(t *testing.T) {
	f := NewDateCheckFilter()
	m, err := f.GetMonth("августа")
	require.NoError(t, err)
	require.Equal(t, 8, m)
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
}

func TestRUDateCheckFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.ru.DateCheckFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
}
