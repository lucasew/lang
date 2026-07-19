package filters

import (
	"fmt"
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
			return h.GetDayOfWeekName(t.Weekday())
		},
		GetMonth: func(localized string) int {
			m, err := h.GetMonth(localized)
			if err != nil {
				return 0
			}
			return int(m)
		},
	}
	return &ArabicDateCheckFilter{AbstractDateCheckFilter: abs, helper: h}
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
func (f *ArabicDateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	wd, err := f.helper.GetDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	return int(wd) + 1, nil
}

func (f *ArabicDateCheckFilter) GetMonth(monthStr string) (int, error) {
	m, err := f.helper.GetMonth(monthStr)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetDayOfWeekName returns Arabic weekday name for a calendar date.
func (f *ArabicDateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return f.helper.GetDayOfWeekName(t.Weekday())
}

// ArabicDMYDateCheckFilter ports org.languagetool.rules.ar.filters.ArabicDMYDateCheckFilter
// as a thin alias (same localization as ArabicDateCheckFilter).
type ArabicDMYDateCheckFilter struct {
	*ArabicDateCheckFilter
}

func NewArabicDMYDateCheckFilter() *ArabicDMYDateCheckFilter {
	return &ArabicDMYDateCheckFilter{ArabicDateCheckFilter: NewArabicDateCheckFilter()}
}

// ValidateDateFilterArgs ensures required keys exist (legacy helper for tests).
func ValidateDateFilterArgs(args map[string]string) error {
	if _, ok := args["weekDay"]; !ok {
		return fmt.Errorf("incomplete args: weekDay required")
	}
	return nil
}
