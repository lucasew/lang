package en

import (
	"strconv"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DateCheckFilter ports org.languagetool.rules.en.DateCheckFilter
// (extends AbstractDateCheckWithSuggestionsFilter with EN localization).
type DateCheckFilter struct {
	helper *DateFilterHelper
	inner  *rules.AbstractDateCheckWithSuggestionsFilter
}

func NewDateCheckFilter() *DateCheckFilter {
	return &DateCheckFilter{
		helper: NewDateFilterHelper(),
		inner:  enDateCheckWithSuggestions(),
	}
}

// GetDayOfWeekJava ports DateCheckFilter.getDayOfWeek(String) via DateFilterHelper.
// Returns Java Calendar day-of-week (Sunday=1 … Saturday=7).
func (f *DateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	wd, err := f.helper.GetDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	// Go time.Weekday: Sunday=0 … Saturday=6 → Java Calendar: Sunday=1 … Saturday=7
	return int(wd) + 1, nil
}

// GetMonth ports DateCheckFilter.getMonth(String) via DateFilterHelper → 1..12.
func (f *DateCheckFilter) GetMonth(monthStr string) (int, error) {
	m, err := f.helper.GetMonth(monthStr)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetDayOfWeekName ports DateCheckFilter.getDayOfWeek(Calendar) via DateFilterHelper
// (Locale.UK LONG display names).
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return f.helper.GetDayOfWeekName(t)
}

// ErrorMessageWrongYear ports DateCheckFilter.getErrorMessageWrongYear.
func (f *DateCheckFilter) ErrorMessageWrongYear() string {
	return `This date is wrong. Did you mean "{currentYear}"?`
}

// DayStrLikeOriginal ports DateCheckFilter.getDayStrLikeOriginal (EN ordinal suffixes).
// Java: if StringTools.isNumeric(original) return day; else append st/nd/rd/th.
func DayStrLikeOriginal(day, original string) string {
	if tools.IsNumeric(original) {
		return day
	}
	number, err := strconv.Atoi(day)
	if err != nil {
		// Java Integer.parseInt would throw; callers always pass numeric day strings.
		return day
	}
	if number >= 11 && number <= 13 {
		return strconv.Itoa(number) + "th"
	}
	switch number % 10 {
	case 1:
		return strconv.Itoa(number) + "st"
	case 2:
		return strconv.Itoa(number) + "nd"
	case 3:
		return strconv.Itoa(number) + "rd"
	default:
		return strconv.Itoa(number) + "th"
	}
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch (super AbstractDateCheckWithSuggestionsFilter).
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if f.inner == nil {
		f.inner = enDateCheckWithSuggestions()
	}
	return f.inner.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
