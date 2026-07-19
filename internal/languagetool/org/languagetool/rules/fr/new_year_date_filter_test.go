package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestNewYearDateFilter(t *testing.T) {
	f := NewNewYearDateFilter()
	jan := true
	y := 2014
	f.ForceJanuary = &jan
	f.ForceYear = &y
	require.True(t, f.ShouldFlag(2013, 2))
	require.False(t, f.ShouldFlag(2014, 2))
	require.False(t, f.ShouldFlag(2013, 12))
	m, err := f.MonthNumber("janvier")
	require.NoError(t, err)
	require.Equal(t, 1, m)
}

func TestNewYearDateFilter_AcceptRuleMatch(t *testing.T) {
	f := NewNewYearDateFilter()
	jan := true
	y := 2014
	f.ForceJanuary = &jan
	f.ForceYear = &y

	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 8, "year {year} should be {realYear}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "3", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "2013")
	require.Contains(t, out.GetMessage(), "2014")

	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "12", "day": "1",
	}, 0, nil, nil))

	// localized month
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "mars", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
}

func TestNewYearDateFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.fr.NewYearDateFilter"))
}
