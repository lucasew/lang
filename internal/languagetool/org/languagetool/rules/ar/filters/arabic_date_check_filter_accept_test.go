package filters

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestArabicDateCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewArabicDateCheckFilter()
	// 27 أغسطس 2014 was Wednesday; claim الجمعة (Friday) → mismatch
	// {realDay} / {day} use ar Locale LONG display names (not helper).
	m := rules.NewRuleMatch(nil, nil, 0, 20, "اليوم {realDay} وليس {day}")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "أغسطس", "day": "27", "weekDay": "الجمعة",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetMessage(), "الأربعاء") // real day
	require.Contains(t, out.GetMessage(), "الجمعة")  // claimed day via Calendar.set(DAY_OF_WEEK)

	// 27 Aug 2014 is Wednesday — correct → null match
	ok := f.AcceptRuleMatch(m, map[string]string{
		"year": "2014", "month": "أغسطس", "day": "27", "weekDay": "الأربعاء",
	}, 0, nil, nil)
	require.Nil(t, ok)

	// Twin-style numeric month: 2022-03-12 Saturday correct / Sunday wrong
	match := rules.NewRuleMatch(nil, nil, 0, 10, "{realDay}/{day}")
	require.Nil(t, f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "السبت",
	}, -1, nil, nil))
	bad := f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "الأحد",
	}, -1, nil, nil)
	require.NotNil(t, bad)
	// real day Saturday = السبت; claimed Sunday calendar display = الأحد
	require.Contains(t, bad.GetMessage(), "السبت")
	require.Contains(t, bad.GetMessage(), "الأحد")

	// Monday displayName path: 2024-01-01 is Monday → ar LONG "الاثنين"
	mon := f.AcceptRuleMatch(m, map[string]string{
		"year": "2024", "month": "1", "day": "1", "weekDay": "الثلاثاء",
	}, 0, nil, nil)
	require.NotNil(t, mon)
	require.Contains(t, mon.GetMessage(), "الاثنين")
}

func TestArabicDateCheckFilter_UnknownWeekdaySkips(t *testing.T) {
	// Java RuntimeException from getDayOfWeek → logger.warn + return null (not crash)
	f := NewArabicDateCheckFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "notaday",
	}, -1, nil, nil)
	require.Nil(t, out)
}

func TestArabicDateCheckFilter_UnknownMonthSkips(t *testing.T) {
	f := NewArabicDateCheckFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"year": "2022", "month": "notamonth", "day": "12", "weekDay": "السبت",
	}, -1, nil, nil)
	require.Nil(t, out)
}

func TestArabicDateCheckFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ar.filters.ArabicDateCheckFilter"))
}
