package it

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter_Helpers(t *testing.T) {
	f := NewDateCheckFilter()
	m, err := f.GetMonth("agosto")
	require.NoError(t, err)
	require.Equal(t, 8, m)
	// 27 agosto 2014 was Wednesday
	require.Equal(t, "mercoledì", f.GetDayOfWeekName(2014, 8, 27))
	jd, err := f.GetDayOfWeekJava("martedì")
	require.NoError(t, err)
	require.Equal(t, 3, jd) // Java Calendar: Sunday=1, Tuesday=3
}

func TestDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewDateCheckFilter()
	// 27 agosto 2014 was Wednesday, not martedì
	m := rules.NewRuleMatch(nil, nil, 0, 20, "Giorno {realDay}, non {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "agosto", "day": "27", "weekDay": "martedì",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "mercoledì")
	require.Contains(t, out.GetMessage(), "martedì")

	// 26 agosto 2014 was Tuesday
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "agosto", "day": "26", "weekDay": "martedì",
	}, 0, nil, nil)
	require.Nil(t, ok)

	// Year optional (grammar rule without year) uses current year — TestMode for 2014
	f.TestMode = true
	// 27 agosto 2014 Wednesday vs lunedì → keep
	out = f.AcceptRuleMatch(m, map[string]string{
		"month": "agosto", "day": "27", "weekDay": "lunedì",
	}, 0, nil, nil)
	require.NotNil(t, out)
}

func TestITDateCheckFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.it.DateCheckFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
}
