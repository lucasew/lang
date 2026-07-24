package eo

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	wd, err := h.GetDayOfWeek("lundo")
	require.NoError(t, err)
	require.Equal(t, 1, int(wd)) // Monday
	m, err := h.GetMonth("aŭgusto")
	require.NoError(t, err)
	require.Equal(t, 8, int(m))
	require.Equal(t, 3, h.GetDayOfMonth("tria"))
	require.Equal(t, 3, h.GetDayOfMonth("trian")) // accusative -n
	require.Equal(t, 16, h.GetDayOfMonth("deksesa"))
	require.Equal(t, 23, h.GetDayOfMonth("dudek-tria"))
}

func TestDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewDateCheckFilter()
	// 27 aŭgusto 2014 was Wednesday
	m := rules.NewRuleMatch(nil, nil, 0, 20, "Estis {realDay}, ne {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "aŭgusto", "day": "27", "weekDay": "lundo",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "merkredo")
	require.Contains(t, out.GetMessage(), "lundo")

	// 26 aŭgusto 2014 was Tuesday = mardo
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "aŭgusto", "day": "26", "weekDay": "mardo",
	}, 0, nil, nil)
	require.Nil(t, ok)
}

func TestEODateCheckFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.eo.DateCheckFilter"))
}
