package ru

import (
	"errors"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FutureDateFilter ports org.languagetool.rules.ru.FutureDateFilter
// (extends AbstractFutureDateFilter; getMonth/getCalendar via DateFilterHelper).
type FutureDateFilter struct {
	helper *DateFilterHelper
	core   *rules.FutureDateFilterCore
}

func NewFutureDateFilter() *FutureDateFilter {
	h := NewDateFilterHelper()
	return &FutureDateFilter{
		helper: h,
		core:   ruFutureDateCoreWith(h),
	}
}

func ruFutureDateCore() *rules.FutureDateFilterCore {
	return ruFutureDateCoreWith(NewDateFilterHelper())
}

// ruFutureDateCoreWith wires AbstractFutureDateFilter hooks via DateFilterHelper
// (Java FutureDateFilter.getMonth / getCalendar).
func ruFutureDateCoreWith(h *DateFilterHelper) *rules.FutureDateFilterCore {
	if h == nil {
		h = NewDateFilterHelper()
	}
	return &rules.FutureDateFilterCore{
		GetMonth: func(localized string) (int, error) {
			// Java DateFilterHelper.getMonth throws RuntimeException on unknown month.
			m, err := h.GetMonth(localized)
			if err != nil {
				return 0, err
			}
			return int(m), nil
		},
	}
}

// GetMonth ports FutureDateFilter.getMonth → DateFilterHelper.getMonth (1–12).
func (f *FutureDateFilter) GetMonth(localizedMonth string) (int, error) {
	if f == nil || f.helper == nil {
		return 0, errNilFutureDateFilter
	}
	m, err := f.helper.GetMonth(localizedMonth)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetCalendar ports FutureDateFilter.getCalendar → DateFilterHelper.getCalendar.
// AbstractFutureDateFilter overlays TestHackHelper 2014-01-01 in acceptRuleMatch
// (implemented in FutureDateFilterCore.now), not here.
func (f *FutureDateFilter) GetCalendar() time.Time {
	if f == nil || f.helper == nil {
		return time.Now()
	}
	return f.helper.GetCalendar()
}

func (f *FutureDateFilter) effectiveCore() *rules.FutureDateFilterCore {
	if f == nil {
		return nil
	}
	if f.core != nil {
		return f.core
	}
	return ruFutureDateCore()
}

// IsFuture reports whether year/month/day (1-based month) is strictly after today.
func (f *FutureDateFilter) IsFuture(year, month, day int) bool {
	return f.effectiveCore().IsFuture(year, month, day)
}

// ParseDayOfMonth extracts leading digits from strings like "23." / "23"
// (AbstractFutureDateFilter DAY_OF_MONTH_PATTERN).
func ParseDayOfMonth(s string) (int, error) {
	return rules.ParseDayOfMonthArg(s, nil)
}

// AcceptRuleMatch ports AbstractFutureDateFilter.acceptRuleMatch:
// keep match only when the date in args is strictly after today
// (invalid calendar → null; TestHackHelper fixed 2014-01-01 under tests).
func (f *FutureDateFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	core := f.effectiveCore()
	if core == nil || !core.AcceptFromArgs(arguments) {
		return nil
	}
	return match
}

// SetNow overrides the reference "today" (tests / calendar inject).
func (f *FutureDateFilter) SetNow(now func() time.Time) {
	if f == nil {
		return
	}
	if f.core == nil {
		f.core = ruFutureDateCore()
	}
	f.core.Now = now
}

var errNilFutureDateFilter = errors.New("nil FutureDateFilter")
