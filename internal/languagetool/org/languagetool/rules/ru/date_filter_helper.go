package ru

import (
	"fmt"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DateFilterHelper ports org.languagetool.rules.ru.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetCalendar ports DateFilterHelper.getCalendar → Calendar.getInstance(Locale.forLanguageTag("ru")).
// Locale selects calendar symbols/first-day; wall clock uses the JVM/OS default TZ
// (Go: time.Now in Local).
func (h *DateFilterHelper) GetCalendar() time.Time {
	return time.Now()
}

// GetDayOfWeek ports DateFilterHelper.getDayOfWeek(String).
// Prefix match after trimSpecialCharacters + toLowerCase (Java), in exact Java order.
// Go time.Weekday Sunday=0 maps to Java Calendar.SUNDAY=1 for the same wall day
// (callers that need Java ints add 1; see DateCheckFilter.GetDayOfWeekJava).
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	// Java: StringTools.trimSpecialCharacters(dayStr).toLowerCase()
	day := strings.ToLower(tools.TrimSpecialCharacters(dayStr))
	switch {
	case strings.HasPrefix(day, "суб"):
		return time.Saturday, nil
	case strings.HasPrefix(day, "вс"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "вос"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "пн"):
		return time.Monday, nil
	case strings.HasPrefix(day, "пон"):
		return time.Monday, nil
	case strings.HasPrefix(day, "вт"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "ср"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "чт"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "чет"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "пт"):
		return time.Friday, nil
	case strings.HasPrefix(day, "пят"):
		return time.Friday, nil
	case strings.HasPrefix(day, "сб"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

// GetDayOfWeekName ports DateFilterHelper.getDayOfWeek(Calendar) —
// Calendar.getDisplayName(DAY_OF_WEEK, LONG, Locale.forLanguageTag("ru")).
func (h *DateFilterHelper) GetDayOfWeekName(t time.Time) string {
	// Java Locale "ru" LONG weekday names (nominative).
	names := []string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"}
	return names[int(t.Weekday())]
}

// GetMonth ports DateFilterHelper.getMonth(String) → 1..12 (as time.Month).
// Prefix match after trimSpecialCharacters + toLowerCase (Java), in exact Java order.
// No Roman numerals / exact full-name equals invent (those live only on DateCheckFilter.java).
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	// Java: StringTools.trimSpecialCharacters(monthStr).toLowerCase()
	mon := strings.ToLower(tools.TrimSpecialCharacters(monthStr))
	switch {
	case strings.HasPrefix(mon, "янв"):
		return time.January, nil
	case strings.HasPrefix(mon, "фев"):
		return time.February, nil
	case strings.HasPrefix(mon, "мар"):
		return time.March, nil
	case strings.HasPrefix(mon, "апр"):
		return time.April, nil
	case strings.HasPrefix(mon, "май"):
		return time.May, nil
	case strings.HasPrefix(mon, "мая"):
		return time.May, nil
	case strings.HasPrefix(mon, "июн"):
		return time.June, nil
	case strings.HasPrefix(mon, "июл"):
		return time.July, nil
	case strings.HasPrefix(mon, "авг"):
		return time.August, nil
	case strings.HasPrefix(mon, "сен"):
		return time.September, nil
	case strings.HasPrefix(mon, "окт"):
		return time.October, nil
	case strings.HasPrefix(mon, "ноя"):
		return time.November, nil
	case strings.HasPrefix(mon, "дек"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
