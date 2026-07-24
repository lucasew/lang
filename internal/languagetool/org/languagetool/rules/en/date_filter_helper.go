package en

import (
	"fmt"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DateFilterHelper ports org.languagetool.rules.en.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetCalendar ports DateFilterHelper.getCalendar → Calendar.getInstance(Locale.UK).
// Locale.UK selects calendar symbols/first-day; wall clock uses the JVM/OS default TZ
// (Go: time.Now in Local).
func (h *DateFilterHelper) GetCalendar() time.Time {
	return time.Now()
}

// GetDayOfWeek ports DateFilterHelper.getDayOfWeek(String).
// Prefix match after trimSpecialCharacters + toLowerCase (Java).
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	// Java: StringTools.trimSpecialCharacters(dayStr).toLowerCase()
	day := strings.ToLower(tools.TrimSpecialCharacters(dayStr))
	switch {
	case strings.HasPrefix(day, "su"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "mo"):
		return time.Monday, nil
	case strings.HasPrefix(day, "tu"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "we"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "th"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "fr"):
		return time.Friday, nil
	case strings.HasPrefix(day, "sa"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

// GetDayOfWeekName ports DateFilterHelper.getDayOfWeek(Calendar) —
// Calendar.getDisplayName(DAY_OF_WEEK, LONG, Locale.UK).
func (h *DateFilterHelper) GetDayOfWeekName(t time.Time) string {
	names := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	return names[int(t.Weekday())]
}

// GetMonth ports DateFilterHelper.getMonth(String) → 1..12 (as time.Month).
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	// Java: StringTools.trimSpecialCharacters(monthStr).toLowerCase()
	mon := strings.ToLower(tools.TrimSpecialCharacters(monthStr))
	switch {
	case strings.HasPrefix(mon, "jan"):
		return time.January, nil
	case strings.HasPrefix(mon, "feb"):
		return time.February, nil
	case strings.HasPrefix(mon, "mar"):
		return time.March, nil
	case strings.HasPrefix(mon, "apr"):
		return time.April, nil
	case strings.HasPrefix(mon, "may"):
		return time.May, nil
	case strings.HasPrefix(mon, "jun"):
		return time.June, nil
	case strings.HasPrefix(mon, "jul"):
		return time.July, nil
	case strings.HasPrefix(mon, "aug"):
		return time.August, nil
	case strings.HasPrefix(mon, "sep"):
		return time.September, nil
	case strings.HasPrefix(mon, "oct"):
		return time.October, nil
	case strings.HasPrefix(mon, "nov"):
		return time.November, nil
	case strings.HasPrefix(mon, "dec"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
