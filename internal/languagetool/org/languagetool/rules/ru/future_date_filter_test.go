package ru

import (
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestFutureDateFilter_IsFuture(t *testing.T) {
	f := NewFutureDateFilter()
	f.SetNow(func() time.Time {
		return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	})
	require.False(t, f.IsFuture(2000, 1, 1))
	require.False(t, f.IsFuture(2014, 1, 1))
	require.True(t, f.IsFuture(2014, 6, 15))
}

func TestParseDayOfMonth(t *testing.T) {
	n, err := ParseDayOfMonth("23.")
	require.NoError(t, err)
	require.Equal(t, 23, n)
}

func TestFutureDateFilter_AcceptRuleMatch(t *testing.T) {
	f := NewFutureDateFilter()
	f.SetNow(func() time.Time {
		return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	})
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 5, "future")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "6", "day": "15",
	}, 0, nil, nil))
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "6", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
	// localized Russian month
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "июнь", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "2", "day": "32",
	}, 0, nil, nil))
}

func TestFutureDateFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.ru.FutureDateFilter"))
}
