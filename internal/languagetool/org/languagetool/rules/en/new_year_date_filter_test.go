package en

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
	require.True(t, f.ShouldFlag(2013, 6))
	require.False(t, f.ShouldFlag(2013, 12)) // December skip
	require.False(t, f.ShouldFlag(2014, 3))  // wrong year
	// EN DateFilterHelper month prefixes
	m, err := f.MonthNumber("January")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("sep")
	require.NoError(t, err)
	require.Equal(t, 9, m)
	// getCalendar via DateFilterHelper (wall clock; non-zero)
	require.False(t, f.GetCalendar().IsZero())
}

func TestNewYearDateFilter_AcceptRuleMatch(t *testing.T) {
	f := NewNewYearDateFilter()
	jan := true
	y := 2014
	f.ForceJanuary = &jan
	f.ForceYear = &y

	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 8, "year {year} should be {realYear}")
	m.SetShortMessage("short")
	m.SetType(rules.RuleMatchTypeHint)

	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "3", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "2013")
	require.Contains(t, out.GetMessage(), "2014")
	require.Equal(t, "short", out.GetShortMessage())
	require.Equal(t, rules.RuleMatchTypeHint, out.GetType())

	// December → null
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "12", "day": "1",
	}, 0, nil, nil))

	// wrong year → null
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2012", "month": "3", "day": "1",
	}, 0, nil, nil))

	// localized EN month via DateFilterHelper
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "March", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)

	// day ordinal suffix (DAY_OF_MONTH_PATTERN)
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "March", "day": "15th",
	}, 0, nil, nil)
	require.NotNil(t, out)

	// invalid calendar day → null
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "2", "day": "30",
	}, 0, nil, nil))

	// soft hyphen on month/day
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "Mar\u00ADch", "day": "1\u00AD5",
	}, 0, nil, nil)
	require.NotNil(t, out)
}

func TestNewYearDateFilter_NotJanuary(t *testing.T) {
	f := NewNewYearDateFilter()
	jan := false
	y := 2014
	f.ForceJanuary = &jan
	f.ForceYear = &y
	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 8, "year {year} should be {realYear}")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "3", "day": "15",
	}, 0, nil, nil))
}

func TestNewYearDateFilter_TestHackUnderGoTest(t *testing.T) {
	// No Force*: IsTest() → January 2014 (Java TestHackHelper)
	f := NewNewYearDateFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 8, "{year}/{realYear}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "June", "day": "1",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, "2013/2014", out.GetMessage())
}

func TestNewYearDateFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.en.NewYearDateFilter"))
}
