package en

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// NewYearDateFilter ports org.languagetool.rules.en.NewYearDateFilter
// (extends AbstractNewYearDateFilter; getMonth/getCalendar via DateFilterHelper).
type NewYearDateFilter struct {
	// dateFilterHelper ports Java private final DateFilterHelper.
	dateFilterHelper *DateFilterHelper
	// ForceJanuary / ForceYear override calendar for tests (Java TestHackHelper).
	ForceJanuary *bool
	ForceYear    *int
	core         *rules.NewYearDateFilterCore
}

func NewNewYearDateFilter() *NewYearDateFilter {
	h := NewDateFilterHelper()
	return &NewYearDateFilter{
		dateFilterHelper: h,
		core:             enNewYearDateCoreWith(h),
	}
}

// GetMonth ports NewYearDateFilter.getMonth → DateFilterHelper.getMonth (1–12).
func (f *NewYearDateFilter) GetMonth(localizedMonth string) (int, error) {
	h := NewDateFilterHelper()
	if f != nil && f.dateFilterHelper != nil {
		h = f.dateFilterHelper
	}
	m, err := h.GetMonth(localizedMonth)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetCalendar ports NewYearDateFilter.getCalendar → DateFilterHelper.getCalendar.
// AbstractNewYearDateFilter overlays TestHackHelper for isJanuary/getCurrentYear.
func (f *NewYearDateFilter) GetCalendar() time.Time {
	if f == nil || f.dateFilterHelper == nil {
		return time.Now()
	}
	return f.dateFilterHelper.GetCalendar()
}

func (f *NewYearDateFilter) effectiveCore() *rules.NewYearDateFilterCore {
	if f == nil {
		return nil
	}
	core := f.core
	if core == nil {
		core = enNewYearDateCore()
	}
	// Copy so Force* on this filter apply without mutating shared defaults.
	c := *core
	c.ForceJanuary = f.ForceJanuary
	c.ForceYear = f.ForceYear
	if c.GetMonth == nil {
		c.GetMonth = core.GetMonth
	}
	if c.GetCalendar == nil {
		c.GetCalendar = f.GetCalendar
	}
	return &c
}

// ShouldFlag reports whether a date with year/month may still refer to the previous
// year wrongly (Java: January + year+1 == currentYear + month != December).
func (f *NewYearDateFilter) ShouldFlag(year, month int) bool {
	return f.effectiveCore().ShouldFlag(year, month)
}

// MonthNumber uses DateFilterHelper (1–12). Twin of getMonth for tests.
func (f *NewYearDateFilter) MonthNumber(localizedMonth string) (int, error) {
	return f.GetMonth(localizedMonth)
}

// AcceptRuleMatch ports AbstractNewYearDateFilter.acceptRuleMatch.
// Keeps match in January for non-December dates with year == currentYear-1;
// rewrites {year} and {realYear} in the message; copies shortMessage and type.
func (f *NewYearDateFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	core := f.effectiveCore()
	if core == nil {
		return nil
	}
	msg := core.AcceptFromArgs(arguments, match.GetMessage())
	if msg == "" {
		return nil
	}
	// Java: new RuleMatch(rule, sentence, from, to, message, shortMessage); setType(match.getType())
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	out.ShortMessage = match.GetShortMessage()
	out.SetType(match.GetType())
	return out
}
