package br

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDateFilterHelper_MonthAndWeekday(t *testing.T) {
	h := NewDateFilterHelper()
	m, err := h.GetMonth("eost")
	require.NoError(t, err)
	require.Equal(t, 8, int(m))
	_, err = h.GetMonth("even")
	require.NoError(t, err)
	wd, err := h.GetDayOfWeek("Dilun") // endsWith lun
	require.NoError(t, err)
	require.Equal(t, 1, int(wd)) // Monday
	wd, err = h.GetDayOfWeek("diriaou")
	require.NoError(t, err)
	require.Equal(t, 4, int(wd)) // Thursday
	require.Equal(t, 1, h.GetDayOfMonth("unan"))
	require.Equal(t, 5, h.GetDayOfMonth("bempvet")) // vet stripped → bemp
	require.Equal(t, 2, h.GetDayOfMonth("taou"))    // soft mutation t→d → daou
}

func TestDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewDateCheckFilter()
	// 27 eost 2014 was Wednesday; claim Lun (Monday) → mismatch
	m := rules.NewRuleMatch(nil, nil, 0, 20, "Said {day} was {realDay}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "27", "weekDay": "Lun",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "Merc’her")
	require.Contains(t, out.GetMessage(), "Lun")

	// 26 eost 2014 was Tuesday = Meurzh
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "eost", "day": "26", "weekDay": "Meurzh",
	}, 0, nil, nil)
	require.Nil(t, ok)
}

func TestBRDateCheckFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.br.DateCheckFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
}
