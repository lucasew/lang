package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestYMDNewYearDateFilter_AcceptRuleMatch_March2013InJan2014(t *testing.T) {
	// Force January 2014 (Java TestHackHelper / Force* equivalents).
	f := NewYMDNewYearDateFilter()
	jan := true
	y := 2014
	f.newYear.ForceJanuary = &jan
	f.newYear.ForceYear = &y

	// date=2013-03-15 → year/month/day via parseDate; {realDate} via correctDate → 2014-03-15;
	// super NewYearDateFilter rewrites {year}/{realYear}.
	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 10, "year {year} → {realYear} date {realDate}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"date":    "2013-03-15",
		"weekDay": "Fr",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "2013")
	require.Contains(t, out.GetMessage(), "2014")
	require.Contains(t, out.GetMessage(), "2014-03-15")
	require.Equal(t, "year 2013 → 2014 date 2014-03-15", out.GetMessage())
}

func TestYMDNewYearDateFilter_AcceptRuleMatch_DecemberSkip(t *testing.T) {
	f := NewYMDNewYearDateFilter()
	jan := true
	y := 2014
	f.newYear.ForceJanuary = &jan
	f.newYear.ForceYear = &y

	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 10, "year {year} → {realYear} date {realDate}")
	// December of previous year: AbstractNewYearDateFilter skips (null).
	out := f.AcceptRuleMatch(m, map[string]string{"date": "2013-12-01", "weekDay": "So"}, 0, nil, nil)
	require.Nil(t, out)
}

func TestYMDNewYearDateFilter_RejectsYearMonthDayKeys(t *testing.T) {
	// Upstream quirk: message uses YMDDateCheckFilter.class.getSimpleName().
	f := NewYMDNewYearDateFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 1, "msg")
	msg := "Set only 'weekDay' and 'date' for YMDDateCheckFilter"
	require.PanicsWithValue(t, msg, func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "2013-03-15", "year": "2013", "weekDay": "Fr"}, 0, nil, nil)
	})
	require.PanicsWithValue(t, msg, func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "2013-03-15", "month": "3", "weekDay": "Fr"}, 0, nil, nil)
	})
	require.PanicsWithValue(t, msg, func() {
		f.AcceptRuleMatch(m, map[string]string{"date": "2013-03-15", "day": "15", "weekDay": "Fr"}, 0, nil, nil)
	})
}

func TestYMDNewYearDateFilter_MissingDatePanics(t *testing.T) {
	f := NewYMDNewYearDateFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 1, "msg")
	require.PanicsWithValue(t, "Missing key 'date'", func() {
		f.AcceptRuleMatch(m, map[string]string{"weekDay": "Fr"}, 0, nil, nil)
	})
}

func TestYMDNewYearDateFilter_TestHackUnderGoTest(t *testing.T) {
	// No Force*: IsTest() → January 2014 (Java TestHackHelper), same as NewYearDateFilter.
	f := NewYMDNewYearDateFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 10, "{year}/{realYear} {realDate}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"date":    "2013-06-01",
		"weekDay": "Sa",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, "2013/2014 2014-06-01", out.GetMessage())
}

func TestYMDNewYearDateFilter_ShortMessageAndType(t *testing.T) {
	// correctDate then super both preserve shortMessage and type (Java RuleMatch copy).
	f := NewYMDNewYearDateFilter()
	jan := true
	y := 2014
	f.newYear.ForceJanuary = &jan
	f.newYear.ForceYear = &y
	m := rules.NewRuleMatch(rules.NewFakeRule("NY"), nil, 0, 10, "{year}/{realYear}")
	m.SetShortMessage("short")
	m.SetType(rules.RuleMatchTypeHint)
	out := f.AcceptRuleMatch(m, map[string]string{"date": "2013-03-15", "weekDay": "Fr"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, "short", out.GetShortMessage())
	require.Equal(t, rules.RuleMatchTypeHint, out.GetType())
}

func TestYMDNewYearDateFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.en.YMDNewYearDateFilter"))
}
