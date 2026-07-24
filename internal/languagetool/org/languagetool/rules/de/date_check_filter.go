package de

import (
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DateCheckFilter ports org.languagetool.rules.de.DateCheckFilter
// (extends AbstractDateCheckWithSuggestionsFilter with DE localization).
type DateCheckFilter struct {
	helper *DateFilterHelper
	inner  *rules.AbstractDateCheckWithSuggestionsFilter
}

func NewDateCheckFilter() *DateCheckFilter {
	return &DateCheckFilter{
		helper: NewDateFilterHelper(),
		inner:  deDateCheckWithSuggestions(),
	}
}

// GetDayOfWeekJava returns Java Calendar day-of-week (Sunday=1 … Saturday=7).
func (f *DateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	wd, err := f.helper.GetDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	// Go: Sunday=0 … Saturday=6 → Java: Sunday=1 … Saturday=7
	return int(wd) + 1, nil
}

// GetMonth returns 1–12.
func (f *DateCheckFilter) GetMonth(monthStr string) (int, error) {
	m, err := f.helper.GetMonth(monthStr)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetDayOfWeekName returns German long weekday name for a date.
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	names := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
	return names[int(t.Weekday())]
}

// AdjustSuggestion ports DateCheckFilter.adjustSuggestion (So. vs Sonntag).
func AdjustDateCheckSuggestion(sugg string) string {
	dotCommaPos := strings.Index(sugg, ".,")
	if dotCommaPos > 5 && dotCommaPos < 12 {
		// remove unnecessary dot
		return strings.Replace(sugg, ".,", ",", 1)
	}
	commaPos := strings.Index(sugg, ",")
	if dotCommaPos < 0 && commaPos > 0 && commaPos < 5 {
		// add dot
		return strings.Replace(sugg, ",", ".,", 1)
	}
	return sugg
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch (super).
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if f.inner == nil {
		f.inner = deDateCheckWithSuggestions()
	}
	return f.inner.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
