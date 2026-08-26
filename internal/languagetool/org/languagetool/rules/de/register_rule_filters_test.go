package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDERuleFiltersRegistered(t *testing.T) {
	for _, class := range []string{
		"org.languagetool.rules.de.FutureDateFilter",
		"org.languagetool.rules.de.DateCheckFilter",
		"org.languagetool.rules.de.NewYearDateFilter",
		"org.languagetool.rules.de.YMDNewYearDateFilter",
		"org.languagetool.rules.de.RemoveUnknownCompoundsFilter",
		"org.languagetool.rules.de.PotentialCompoundFilter",
		"org.languagetool.rules.de.CompoundCheckFilter",
		"org.languagetool.rules.de.InsertCommaFilter",
		"org.languagetool.rules.de.RecentYearFilter",
		"org.languagetool.rules.de.ValidWordFilter",
	} {
		require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
		require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(class), class)
	}
}

func TestDEFutureDateFilter(t *testing.T) {
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.de.FutureDateFilter")
	m := rules.NewRuleMatch(nil, nil, 0, 5, "msg")
	// far future → keep
	out := f.AcceptRuleMatch(m, map[string]string{"year": "2099", "month": "März", "day": "15"}, 0, nil, nil)
	require.NotNil(t, out)
	// past → drop
	out = f.AcceptRuleMatch(m, map[string]string{"year": "2000", "month": "März", "day": "15"}, 0, nil, nil)
	require.Nil(t, out)
}
