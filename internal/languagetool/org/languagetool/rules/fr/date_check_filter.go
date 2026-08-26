package fr

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DateCheckFilter ports org.languagetool.rules.fr.DateCheckFilter
// (extends AbstractDateCheckWithSuggestionsFilter with FR localization).
type DateCheckFilter struct {
	helper *DateFilterHelper
	inner  *rules.AbstractDateCheckWithSuggestionsFilter
}

func NewDateCheckFilter() *DateCheckFilter {
	return &DateCheckFilter{
		helper: NewDateFilterHelper(),
		inner:  frDateCheckWithSuggestions(),
	}
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
	return formatFRDayOfWeek(t)
}

func formatFRDayOfWeek(t time.Time) string {
	names := []string{"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi"}
	return names[int(t.Weekday())]
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch (super).
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if f.inner == nil {
		f.inner = frDateCheckWithSuggestions()
	}
	return f.inner.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
