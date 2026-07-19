package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestYMDNewYearDateFilter(t *testing.T) {
	f := NewYMDNewYearDateFilter()
	jan := true
	y := 2014
	f.newYear.ForceJanuary = &jan
	f.newYear.ForceYear = &y
	ok, err := f.ShouldFlagFromArgs(map[string]string{"date": "2013-03-15", "weekDay": "Fr"})
	require.NoError(t, err)
	require.True(t, ok)
	_, err = f.PrepareArgs(map[string]string{"date": "2013-03-15", "year": "2013"})
	require.Error(t, err)
}

func TestYMDNewYearDateFilter_AcceptRuleMatch(t *testing.T) {
	f := NewYMDNewYearDateFilter()
	jan := true
	y := 2014
	f.newYear.ForceJanuary = &jan
	f.newYear.ForceYear = &y

	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 10, "year {year} → {realYear} date {realDate}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"date":    "2013-03-15",
		"weekDay": "Fr",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "2013")
	require.Contains(t, out.GetMessage(), "2014")
	require.Contains(t, out.GetMessage(), "2014-03-15")

	// December of previous year: not flagged
	out = f.AcceptRuleMatch(m, map[string]string{"date": "2013-12-01", "weekDay": "So"}, 0, nil, nil)
	require.Nil(t, out)
}

func TestYMDNewYearDateFilter_RejectsYearKey(t *testing.T) {
	f := NewYMDNewYearDateFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 1, "msg")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "2013-03-15", "year": "2013", "weekDay": "Fr"}, 0, nil, nil)
	})
}

func TestYMDNewYearDateFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.de.YMDNewYearDateFilter"))
}
