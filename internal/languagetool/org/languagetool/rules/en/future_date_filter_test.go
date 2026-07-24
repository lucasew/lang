package en

import (
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestFutureDateFilter(t *testing.T) {
	f := NewFutureDateFilter()
	f.SetNow(func() time.Time {
		return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	})
	require.False(t, f.IsFuture(1999, 12, 31))
	require.False(t, f.IsFuture(2014, 1, 1)) // not strictly after
	require.True(t, f.IsFuture(2014, 6, 15))
	d, err := ParseDayOfMonth("23rd")
	require.NoError(t, err)
	require.Equal(t, 23, d)
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
	// localized English month
	out = f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "June", "day": "15",
	}, 0, nil, nil)
	require.NotNil(t, out)
	// invalid calendar date → drop
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"year": "2015", "month": "2", "day": "32",
	}, 0, nil, nil))
}

func TestFutureDateFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.en.FutureDateFilter"))
}
