package filters

import (
	"fmt"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicDateCheckFilter ports org.languagetool.rules.ar.filters.ArabicDateCheckFilter.
type ArabicDateCheckFilter struct {
	helper *ArabicDateFilterHelper
	base   *rules.AbstractDateCheckFilter
}

func NewArabicDateCheckFilter() *ArabicDateCheckFilter {
	h := NewArabicDateFilterHelper()
	f := &ArabicDateCheckFilter{helper: h}
	f.base = &rules.AbstractDateCheckFilter{
		GetDayOfWeekName: func(localized string) time.Weekday {
			wd, err := h.GetDayOfWeek(localized)
			if err != nil {
				return time.Sunday
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
	return f
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

func (f *ArabicDateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return f.helper.GetDayOfWeekName(t.Weekday())
}

// AcceptRuleMatch delegates to AbstractDateCheckFilter.
func (f *ArabicDateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, args map[string]string) *rules.RuleMatch {
	if f == nil || f.base == nil {
		return nil
	}
	return f.base.AcceptRuleMatch(match, args)
}

// DMYArabicDateCheckFilter is a thin alias for day-month-year oriented checks.
type ArabicDMYDateCheckFilter struct {
	*ArabicDateCheckFilter
}

func NewArabicDMYDateCheckFilter() *ArabicDMYDateCheckFilter {
	return &ArabicDMYDateCheckFilter{ArabicDateCheckFilter: NewArabicDateCheckFilter()}
}

// ValidateArgs ensures required keys exist (pattern-rule filter contract).
func ValidateDateFilterArgs(args map[string]string) error {
	if _, ok := args["weekDay"]; !ok {
		return fmt.Errorf("incomplete args: weekDay required")
	}
	return nil
}
