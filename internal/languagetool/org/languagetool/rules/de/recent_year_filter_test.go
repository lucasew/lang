package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRecentYearFilter(t *testing.T) {
	f := NewRecentYearFilter()
	y := 2020
	f.ForceYear = &y
	// Java: year < thisYear && year >= thisYear-maxYearsBack
	require.True(t, f.Accept(2019, 5))
	require.True(t, f.Accept(2015, 5)) // thisYear-5 inclusive
	require.False(t, f.Accept(2014, 5))
	require.False(t, f.Accept(2020, 5)) // current year excluded
}

func TestRecentYearFilter_AcceptRuleMatch(t *testing.T) {
	y := 2020
	f := NewRecentYearFilter()
	f.ForceYear = &y
	m := rules.NewRuleMatch(rules.NewFakeRule("Y"), nil, 0, 4, "msg")
	require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{"year": "2019", "maxYearsBack": "5"}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"year": "2010", "maxYearsBack": "5"}, 0, nil, nil))
	// bad args: Java Integer.parseInt throws; Go fail-closed nil (no invent)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"year": "x", "maxYearsBack": "5"}, 0, nil, nil))
}

func TestRecentYearFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.de.RecentYearFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.de.RecentYearFilter"))
}
