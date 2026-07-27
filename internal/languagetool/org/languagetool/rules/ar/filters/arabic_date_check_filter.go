package filters

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicDateCheckFilter ports org.languagetool.rules.ar.filters.ArabicDateCheckFilter
// (extends AbstractDateCheckFilter).
type ArabicDateCheckFilter struct {
	*rules.AbstractDateCheckFilter
	helper *ArabicDateFilterHelper
}

func NewArabicDateCheckFilter() *ArabicDateCheckFilter {
	h := NewArabicDateFilterHelper()
	f := &ArabicDateCheckFilter{helper: h}
	abs := &rules.AbstractDateCheckFilter{
		GetDayOfWeekName: func(localized string) time.Weekday {
			wd, err := h.GetDayOfWeek(localized)
			if err != nil {
				// Java throws RuntimeException on unknown weekday → AbstractDateCheckFilter
				// catches RuntimeException (not IllegalArgumentException) and returns null.
				panic(err)
			}
			return wd
		},
		// Java ArabicDateCheckFilter.getDayOfWeek(Calendar):
		// date.getDisplayName(Calendar.DAY_OF_WEEK, Calendar.LONG, Locale.forLanguageTag("ar"))
		// NOT helper.getDayOfWeekName (that is the int overload only).
		FormatDayOfWeek: func(t time.Time) string {
			return arabicLocaleLongDayOfWeek(t)
		},
		GetMonth: func(localized string) int {
			m, err := h.GetMonth(localized)
			if err != nil {
				// Java throws RuntimeException on unknown month.
				panic(err)
			}
			return int(m)
		},
	}
	f.AbstractDateCheckFilter = abs
	return f
}

// arabicLocaleLongDayOfWeek ports
// Calendar.getDisplayName(DAY_OF_WEEK, LONG, Locale.forLanguageTag("ar")).
// Verified against OpenJDK Calendar (Monday = "الاثنين" without hamza on alif;
// helper getDayOfWeekName returns "الإثنين" with hamza — different path).
func arabicLocaleLongDayOfWeek(t time.Time) string {
	// index = Go time.Weekday: Sunday=0 … Saturday=6
	names := []string{"الأحد", "الاثنين", "الثلاثاء", "الأربعاء", "الخميس", "الجمعة", "السبت"}
	return names[int(t.Weekday())]
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch via AbstractDateCheckFilter.
func (f *ArabicDateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractDateCheckFilter == nil {
		return nil
	}
	return f.AbstractDateCheckFilter.AcceptRuleMatch(match, arguments)
}

// GetDayOfWeekJava returns Java Calendar day-of-week (Sunday=1 … Saturday=7).
// Ports ArabicDateCheckFilter.getDayOfWeek(String) via helper.
func (f *ArabicDateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	wd, err := f.helper.GetDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	// Go Sunday=0 … Saturday=6 → Java Sunday=1 … Saturday=7
	return int(wd) + 1, nil
}

// GetMonth ports ArabicDateCheckFilter.getMonth(String) → 1..12 via helper.
func (f *ArabicDateCheckFilter) GetMonth(monthStr string) (int, error) {
	m, err := f.helper.GetMonth(monthStr)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetDayOfWeekName ports ArabicDateCheckFilter.getDayOfWeek(Calendar) —
// Locale.forLanguageTag("ar") LONG display names (not helper names).
func (f *ArabicDateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return arabicLocaleLongDayOfWeek(t)
}

// GetDayOfWeekNameFromJavaDay ports package-visible ArabicDateCheckFilter.getDayOfWeek(int)
// via helper.getDayOfWeekName (returns "الإثنين" for Monday — distinct from Calendar displayName).
func (f *ArabicDateCheckFilter) GetDayOfWeekNameFromJavaDay(javaDay int) string {
	// Java Calendar: SUNDAY=1 … SATURDAY=7
	return f.helper.GetDayOfWeekName(time.Weekday(javaDay - 1))
}
