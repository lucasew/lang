package ru

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DateCheckFilter ports org.languagetool.rules.ru.DateCheckFilter
// (extends AbstractDateCheckFilter).
// Localization day/month helpers currently share DateFilterHelper (FutureDateFilter path);
// Java DateCheckFilter still has its own older equals/Roman getMonth — separate sector.
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
				// Java throws RuntimeException on unknown weekday.
				panic(err)
			}
			return wd
		},
		FormatDayOfWeek: func(t time.Time) string {
			// DateFilterHelper.getDayOfWeek(Calendar) — Locale.ru LONG
			return h.GetDayOfWeekName(t)
		},
		GetMonth: func(localized string) int {
			m, err := h.GetMonth(localized)
			if err != nil {
				return 0
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
func (f *DateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	wd, err := f.helper.GetDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	return int(wd) + 1, nil
}

func (f *DateCheckFilter) GetMonth(monthStr string) (int, error) {
	m, err := f.helper.GetMonth(monthStr)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetDayOfWeekName returns Russian weekday name for a calendar date.
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return f.helper.GetDayOfWeekName(t)
}
