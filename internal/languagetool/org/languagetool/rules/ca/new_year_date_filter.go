package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// NewYearDateFilter ports org.languagetool.rules.ca.NewYearDateFilter
// (extends AbstractNewYearDateFilter with CA month localization).
type NewYearDateFilter struct {
	// ForceJanuary / ForceYear override calendar for tests (Java TestHackHelper).
	ForceJanuary *bool
	ForceYear    *int
	core         *rules.NewYearDateFilterCore
}

func NewNewYearDateFilter() *NewYearDateFilter {
	return &NewYearDateFilter{
		core: caNewYearDateCore(),
	}
}

func (f *NewYearDateFilter) effectiveCore() *rules.NewYearDateFilterCore {
	if f == nil {
		return nil
	}
	core := f.core
	if core == nil {
		core = caNewYearDateCore()
	}
	c := *core
	c.ForceJanuary = f.ForceJanuary
	c.ForceYear = f.ForceYear
	if c.GetMonth == nil {
		c.GetMonth = core.GetMonth
	}
	return &c
}

// ShouldFlag reports whether a date with year/month may still refer to the previous
// year wrongly (Java: January + year+1 == currentYear + month != December).
func (f *NewYearDateFilter) ShouldFlag(year, month int) bool {
	return f.effectiveCore().ShouldFlag(year, month)
}

// MonthNumber uses DateFilterHelper (1–12).
func (f *NewYearDateFilter) MonthNumber(localizedMonth string) (int, error) {
	core := f.effectiveCore()
	if core != nil && core.GetMonth != nil {
		return core.GetMonth(localizedMonth)
	}
	m, err := NewDateFilterHelper().GetMonth(localizedMonth)
	return int(m), err
}

// AcceptRuleMatch ports AbstractNewYearDateFilter.acceptRuleMatch.
// Keeps match in January for non-December dates with year == currentYear-1;
// rewrites {year} and {realYear} in the message.
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
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	out.ShortMessage = match.ShortMessage
	out.IssueType = match.IssueType
	return out
}
