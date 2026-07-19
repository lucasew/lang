package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestNewYearDateFilter_ShouldFlag(t *testing.T) {
	jan := true
	y := 2024
	f := &NewYearDateFilter{ForceJanuary: &jan, ForceYear: &y, core: caNewYearDateCore()}
	require.True(t, f.ShouldFlag(2023, 1))
	require.False(t, f.ShouldFlag(2023, 12))
	require.False(t, f.ShouldFlag(2024, 1))
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

	// localized Catalan month if helper supports it
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "març", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
}

func TestNewYearDateFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.ca.NewYearDateFilter"))
}
