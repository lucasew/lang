package pt

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FutureDateFilter ports org.languagetool.rules.pt.FutureDateFilter
// (extends AbstractFutureDateFilter with PT month localization).
type FutureDateFilter struct {
	helper *DateFilterHelper
	core   *rules.FutureDateFilterCore
}

func NewFutureDateFilter() *FutureDateFilter {
	return &FutureDateFilter{
		helper: NewDateFilterHelper(),
		core:   ptFutureDateCore(),
	}
}

func (f *FutureDateFilter) effectiveCore() *rules.FutureDateFilterCore {
	if f == nil {
		return nil
	}
	if f.core != nil {
		return f.core
	}
	return ptFutureDateCore()
}

// IsFuture reports whether year/month/day (1-based month) is strictly after today.
func (f *FutureDateFilter) IsFuture(year, month, day int) bool {
	return f.effectiveCore().IsFuture(year, month, day)
}

// ParseDayOfMonth extracts leading digits from strings like "23." / "23".
func ParseDayOfMonth(s string) (int, error) {
	return rules.ParseDayOfMonthArg(s, nil)
}

// AcceptRuleMatch ports AbstractFutureDateFilter.acceptRuleMatch:
// keep match only when the date in args is strictly after today.
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
		f.core = ptFutureDateCore()
	}
	f.core.Now = now
}
