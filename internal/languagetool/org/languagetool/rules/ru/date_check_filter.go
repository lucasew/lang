package ru

import (
	"fmt"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DateCheckFilter ports org.languagetool.rules.ru.DateCheckFilter
// (extends AbstractDateCheckFilter).
//
// Day/month localization is implemented here exactly as DateCheckFilter.java —
// NOT via DateFilterHelper (which has different startsWith/trim logic and no Romans).
type DateCheckFilter struct {
	*rules.AbstractDateCheckFilter
}

func NewDateCheckFilter() *DateCheckFilter {
	f := &DateCheckFilter{}
	abs := &rules.AbstractDateCheckFilter{
		GetDayOfWeekName: func(localized string) time.Weekday {
			wd, err := f.getDayOfWeek(localized)
			if err != nil {
				// Java throws RuntimeException on unknown weekday.
				panic(err)
			}
			return wd
		},
		FormatDayOfWeek: func(t time.Time) string {
			return f.getDayOfWeekName(t)
		},
		GetMonth: func(localized string) int {
			m, err := f.getMonth(localized)
			if err != nil {
				// Java throws RuntimeException on unknown month.
				panic(err)
			}
			return m
		},
	}
	f.AbstractDateCheckFilter = abs
	return f
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch via AbstractDateCheckFilter.
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractDateCheckFilter == nil {
		return nil
	}
	return f.AbstractDateCheckFilter.AcceptRuleMatch(match, arguments)
}

// getDayOfWeek ports DateCheckFilter.getDayOfWeek(String).
// No trimSpecialCharacters (unlike DateFilterHelper); only toLowerCase.
//
//	пн OR equals понедельник → MONDAY
//	вт → TUESDAY
//	ср → WEDNESDAY
//	чт OR equals четверг → THURSDAY
//	equals пт OR startsWith пятниц → FRIDAY
//	startsWith сб OR startsWith суббот → SATURDAY
//	startsWith вс OR equals воскресенье → SUNDAY
func (f *DateCheckFilter) getDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasPrefix(day, "пн") || day == "понедельник":
		return time.Monday, nil
	case strings.HasPrefix(day, "вт"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "ср"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "чт") || day == "четверг":
		return time.Thursday, nil
	case day == "пт" || strings.HasPrefix(day, "пятниц"):
		return time.Friday, nil
	case strings.HasPrefix(day, "сб") || strings.HasPrefix(day, "суббот"):
		return time.Saturday, nil
	case strings.HasPrefix(day, "вс") || day == "воскресенье":
		return time.Sunday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for '%s'", dayStr)
	}
}

// getDayOfWeekName ports DateCheckFilter.getDayOfWeek(Calendar) —
// date.getDisplayName(DAY_OF_WEEK, LONG, Locale.forLanguageTag("ru")).
func (f *DateCheckFilter) getDayOfWeekName(t time.Time) string {
	// Java Locale "ru" LONG weekday names (nominative).
	names := []string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"}
	return names[int(t.Weekday())]
}

// getMonth ports DateCheckFilter.getMonth(String) → 1..12.
// Russian names compared after mon=monthStr.toLowerCase(); Roman numerals
// use original monthStr.equals (case-sensitive I..XII).
func (f *DateCheckFilter) getMonth(monthStr string) (int, error) {
	mon := strings.ToLower(monthStr)
	switch {
	case mon == "январь" || monthStr == "I" || mon == "января" || mon == "янв":
		return 1, nil
	case mon == "февраль" || monthStr == "II" || mon == "февраля" || mon == "фев":
		return 2, nil
	case mon == "март" || monthStr == "III" || mon == "марта" || mon == "мар":
		return 3, nil
	case mon == "апрель" || monthStr == "IV" || mon == "апреля" || mon == "апр":
		return 4, nil
	case mon == "май" || monthStr == "V" || mon == "мая":
		return 5, nil
	case mon == "июнь" || monthStr == "VI" || mon == "июня" || mon == "ин":
		return 6, nil
	case mon == "июль" || monthStr == "VII" || mon == "июля" || mon == "ил":
		return 7, nil
	case mon == "август" || monthStr == "VIII" || mon == "августа" || mon == "авг":
		return 8, nil
	case mon == "сентябрь" || monthStr == "IX" || mon == "сентября" || mon == "сен":
		return 9, nil
	case mon == "октябрь" || monthStr == "X" || mon == "октября" || mon == "окт":
		return 10, nil
	case mon == "ноябрь" || monthStr == "XI" || mon == "ноября" || mon == "ноя":
		return 11, nil
	case mon == "декабрь" || monthStr == "XII" || mon == "декабря" || mon == "дек":
		return 12, nil
	default:
		return 0, fmt.Errorf("could not find month '%s'", monthStr)
	}
}

// GetDayOfWeekJava returns Java Calendar day-of-week (Sunday=1 … Saturday=7).
func (f *DateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	if f == nil {
		return 0, fmt.Errorf("nil DateCheckFilter")
	}
	wd, err := f.getDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	// Go Sunday=0 … Saturday=6 → Java Sunday=1 … Saturday=7
	return int(wd) + 1, nil
}

// GetMonth ports DateCheckFilter.getMonth for tests (1=January … 12=December).
func (f *DateCheckFilter) GetMonth(monthStr string) (int, error) {
	if f == nil {
		return 0, fmt.Errorf("nil DateCheckFilter")
	}
	return f.getMonth(monthStr)
}

// GetDayOfWeekName returns Russian LONG weekday name for a calendar date.
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return f.getDayOfWeekName(t)
}
