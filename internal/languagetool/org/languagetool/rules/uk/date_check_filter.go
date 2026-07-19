package uk

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// DateCheckFilter ports org.languagetool.rules.uk.DateCheckFilter
// (extends AbstractDateCheckFilter — not WithSuggestions — with UK localization).
type DateCheckFilter struct {
	helper *DateFilterHelper
	inner  *rules.AbstractDateCheckFilter
}

func NewDateCheckFilter() *DateCheckFilter {
	return &DateCheckFilter{
		helper: NewDateFilterHelper(),
		inner:  ukDateCheckFilter(),
	}
}

func ukDateCheckFilter() *rules.AbstractDateCheckFilter {
	h := NewDateFilterHelper()
	return &rules.AbstractDateCheckFilter{
		GetDayOfWeekName: func(localized string) time.Weekday {
			wd, err := h.GetDayOfWeek(localized)
			if err != nil {
				panic(err)
			}
			return wd
		},
		FormatDayOfWeek: formatUKDayOfWeek,
		GetMonth: func(localized string) int {
			m, err := h.GetMonth(localized)
			if err != nil {
				return 0
			}
			return int(m)
		},
	}
}

func formatUKDayOfWeek(t time.Time) string {
	// Java: Calendar LONG display name for Locale "uk"
	names := []string{"неділя", "понеділок", "вівторок", "середа", "четвер", "пʼятниця", "субота"}
	return names[int(t.Weekday())]
}

func init() {
	patterns.GlobalRuleFilterCreator.Register("org.languagetool.rules.uk.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
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

func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return formatUKDayOfWeek(t)
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch (super AbstractDateCheckFilter).
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if f.inner == nil {
		f.inner = ukDateCheckFilter()
	}
	return f.inner.AcceptRuleMatch(match, arguments)
}
