package en

import (
	"strconv"
	"time"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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
	names := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	return names[int(t.Weekday())]
}

// DayStrLikeOriginal ports DateCheckFilter.getDayStrLikeOriginal (EN ordinal suffixes).
func DayStrLikeOriginal(day, original string) string {
	if isNumeric(original) {
		return day
	}
	number, err := strconv.Atoi(day)
	if err != nil {
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

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch (super).
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
