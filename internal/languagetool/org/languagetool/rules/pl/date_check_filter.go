package pl

import (
	"fmt"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DateCheckFilter ports org.languagetool.rules.pl.DateCheckFilter
// (extends AbstractDateCheckFilter — not WithSuggestions).
//
// Java has no separate DateFilterHelper; day/month localization lives on this type.
// Go keeps DateFilterHelper as the localization carrier (same pattern as EO).
type DateCheckFilter struct {
	*rules.AbstractDateCheckFilter
	helper *DateFilterHelper
}

func NewDateCheckFilter() *DateCheckFilter {
	h := NewDateFilterHelper()
	abs := &rules.AbstractDateCheckFilter{
		GetDayOfWeekName: func(localized string) time.Weekday {
			wd, err := h.GetDayOfWeek(localized)
			if err != nil {
				// Java DateCheckFilter throws RuntimeException on unknown weekday.
				panic(err)
			}
			return wd
		},
		// Java: date.getDisplayName(DAY_OF_WEEK, LONG, Locale.forLanguageTag("pl"))
		FormatDayOfWeek: func(t time.Time) string {
			return h.GetDayOfWeekName(t)
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
	return &DateCheckFilter{AbstractDateCheckFilter: abs, helper: h}
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch via AbstractDateCheckFilter.
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractDateCheckFilter == nil {
		return nil
	}
	return f.AbstractDateCheckFilter.AcceptRuleMatch(match, arguments)
}

// GetDayOfWeekJava returns Java Calendar day-of-week (Sunday=1 … Saturday=7).
// Ports DateCheckFilter.getDayOfWeek(String) for tests.
func (f *DateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	if f == nil || f.helper == nil {
		return 0, fmt.Errorf("nil DateCheckFilter")
	}
	wd, err := f.helper.GetDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	// Go Sunday=0 … Saturday=6 → Java Sunday=1 … Saturday=7
	return int(wd) + 1, nil
}

// GetMonth ports DateCheckFilter.getMonth(String) → 1..12 for tests.
func (f *DateCheckFilter) GetMonth(monthStr string) (int, error) {
	if f == nil || f.helper == nil {
		return 0, fmt.Errorf("nil DateCheckFilter")
	}
	m, err := f.helper.GetMonth(monthStr)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetDayOfWeekName returns Polish LONG weekday name for a calendar date.
// Ports DateCheckFilter.getDayOfWeek(Calendar).
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return f.helper.GetDayOfWeekName(t)
}
