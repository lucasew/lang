package ru

import (
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Fixed "today" matching Java TestHackHelper / Calendar.Builder.setDate(2014, 0, 1).
func fixed2014() time.Time {
	return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.Local)
}

func TestFutureDateFilter_IsFuture(t *testing.T) {
	f := NewFutureDateFilter()
	f.SetNow(fixed2014)
	// Java dateFromDate.after(currentDate) — strictly after 2014-01-01
	require.False(t, f.IsFuture(1999, 12, 31))
	require.False(t, f.IsFuture(2014, 1, 1), "same day is not after")
	require.True(t, f.IsFuture(2014, 1, 2))
	require.True(t, f.IsFuture(2014, 6, 15))
	// invalid calendar (lenient false) → not future
	require.False(t, f.IsFuture(2015, 2, 30))
	require.False(t, f.IsFuture(2015, 2, 32))
}

func TestFutureDateFilter_ParseDayOfMonth(t *testing.T) {
	// DAY_OF_MONTH_PATTERN (\\d+).* full Matcher.matches — RU often "23."
	d, err := ParseDayOfMonth("23.")
	require.NoError(t, err)
	require.Equal(t, 23, d)
	d, err = ParseDayOfMonth("1.")
	require.NoError(t, err)
	require.Equal(t, 1, d)
	d, err = ParseDayOfMonth("15")
	require.NoError(t, err)
	require.Equal(t, 15, d)
	// no full match → Java getDayOfMonth default 0 (RU does not override)
	d, err = ParseDayOfMonth("x22")
	require.NoError(t, err)
	require.Equal(t, 0, d)
}

func TestFutureDateFilter_GetMonth(t *testing.T) {
	f := NewFutureDateFilter()
	// DateFilterHelper prefixes (trimSpecialCharacters + lower) — no Romans
	m, err := f.GetMonth("июнь")
	require.NoError(t, err)
	require.Equal(t, 6, m)
	m, err = f.GetMonth("янв")
	require.NoError(t, err)
	require.Equal(t, 1, m)
	m, err = f.GetMonth("авг")
	require.NoError(t, err)
	require.Equal(t, 8, m)
	m, err = f.GetMonth("августа")
	require.NoError(t, err)
	require.Equal(t, 8, m)
	m, err = f.GetMonth("декабрь")
	require.NoError(t, err)
	require.Equal(t, 12, m)
	_, err = f.GetMonth("notamonth")
	require.Error(t, err)
}

func TestFutureDateFilter_GetCalendar(t *testing.T) {
	f := NewFutureDateFilter()
	// DateFilterHelper.getCalendar → wall clock (not TestHack 2014)
	before := time.Now().Add(-time.Minute)
	cal := f.GetCalendar()
	after := time.Now().Add(time.Minute)
	require.True(t, !cal.Before(before) && !cal.After(after))
}

func TestFutureDateFilter_AcceptRuleMatch(t *testing.T) {
	f := NewFutureDateFilter()
	f.SetNow(fixed2014)
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 5, "future")

	// past → drop
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "6", "day": "15",
	}, 0, nil, nil))
	// same day → drop (not strictly after)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "1", "day": "1",
	}, 0, nil, nil))
	// future numeric month → keep
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "6", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, m, out)
	// localized Russian month name (DateFilterHelper prefix "июн")
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "июнь", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
	// abbreviated month
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "авг", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
	// genitive form (prefix "авг")
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "августа", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
	// day with trailing period (DAY_OF_MONTH_PATTERN "23.")
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "6", "day": "15.",
	}, 0, nil, nil)
	require.NotNil(t, out)
	// invalid calendar date → drop (Java IllegalArgumentException → null)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "2", "day": "32",
	}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "2", "day": "30",
	}, 0, nil, nil))
	// non-numeric day without RU getDayOfMonth → day 0 → invalid → drop
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "6", "day": "пятнадцатое",
	}, 0, nil, nil))
}

func TestFutureDateFilter_AcceptRuleMatch_MissingKeys(t *testing.T) {
	f := NewFutureDateFilter()
	f.SetNow(fixed2014)
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 5, "future")
	// Java getRequired → IllegalArgumentException "Missing key '…'"
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"month": "6", "day": "15"}, 0, nil, nil)
	})
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"year": "2015", "day": "15"}, 0, nil, nil)
	})
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"year": "2015", "month": "6"}, 0, nil, nil)
	})
}

func TestFutureDateFilter_TestHackHelperDefaultNow(t *testing.T) {
	// Under go test, IsTest() is true → currentDate = 2014-01-01 without SetNow.
	f := NewFutureDateFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 5, "future")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2013", "month": "12", "day": "31",
	}, 0, nil, nil))
	require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "6", "day": "1",
	}, 0, nil, nil))
	// IsFuture also uses TestHackHelper when Now unset
	require.True(t, f.IsFuture(2014, 6, 1))
	require.False(t, f.IsFuture(2013, 6, 1))
}

func TestFutureDateFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.ru.FutureDateFilter"))
	// factory returns AcceptRuleMatch-capable filter
	rf := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.ru.FutureDateFilter")
	require.NotNil(t, rf)
	m := rules.NewRuleMatch(rules.NewFakeRule("F"), nil, 0, 5, "future")
	// past relative to TestHackHelper 2014-01-01
	require.Nil(t, rf.AcceptRuleMatch(m, map[string]string{
		"year": "2000", "month": "1", "day": "1",
	}, 0, nil, nil))
	require.NotNil(t, rf.AcceptRuleMatch(m, map[string]string{
		"year": "2099", "month": "декабрь", "day": "31",
	}, 0, nil, nil))
}
