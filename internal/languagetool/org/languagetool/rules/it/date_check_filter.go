package it

import (
	"fmt"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DateCheckFilter ports org.languagetool.rules.it.DateCheckFilter
// (extends AbstractDateCheckFilter).
//
// Java does not override getDayOfMonth — default returns 0 (no ordinal invent).
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
				// Java throws RuntimeException on unknown weekday → AbstractDateCheckFilter
				// catches RuntimeException (not IllegalArgumentException) and returns null.
				panic(err)
			}
			return wd
		},
		// Java DateCheckFilter.getDayOfWeek(Calendar):
		// englishDay = date.getDisplayName(DAY_OF_WEEK, LONG, Locale.UK) then map to Italian.
		FormatDayOfWeek: func(t time.Time) string {
			return itDayOfWeekFromUKLong(t)
		},
		GetMonth: func(localized string) int {
			m, err := h.GetMonth(localized)
			if err != nil {
				// Java throws RuntimeException on unknown month.
				panic(err)
			}
			return int(m)
		},
		// No GetDayOfMonthOptional: IT DateCheckFilter does not override getDayOfMonth.
	}
	return &DateCheckFilter{AbstractDateCheckFilter: abs, helper: h}
}

// itDayOfWeekFromUKLong ports DateCheckFilter.getDayOfWeek(Calendar):
// Locale.UK LONG English → Italian names (Java if-chain on EnglishDay).
func itDayOfWeekFromUKLong(t time.Time) string {
	switch t.Weekday() {
	case time.Sunday:
		return "domenica"
	case time.Monday:
		return "lunedì"
	case time.Tuesday:
		return "martedì"
	case time.Wednesday:
		return "mercoledì"
	case time.Thursday:
		return "giovedì"
	case time.Friday:
		return "venerdì"
	case time.Saturday:
		return "sabato"
	default:
		return ""
	}
}

// GetDayOfWeekJava returns Java Calendar day-of-week (Sunday=1 … Saturday=7).
// Ports DateCheckFilter.getDayOfWeek(String).
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

// GetDayOfWeekName ports DateCheckFilter.getDayOfWeek(Calendar) for a civil date.
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return itDayOfWeekFromUKLong(t)
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch via AbstractDateCheckFilter.
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractDateCheckFilter == nil {
		return nil
	}
	return f.AbstractDateCheckFilter.AcceptRuleMatch(match, arguments)
}
