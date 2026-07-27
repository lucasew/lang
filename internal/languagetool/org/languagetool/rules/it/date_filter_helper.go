package it

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper ports Italian DateCheckFilter day/month localization methods
// (no separate Java type; methods live on org.languagetool.rules.it.DateCheckFilter).
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetDayOfWeek ports DateCheckFilter.getDayOfWeek(String).
// Prefix / exact equals order matches Java sequential checks.
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	// Java: if (day.startsWith("do") || day.equals("domenica")) …
	// equals branches are redundant with startsWith for full names but kept bug-for-bug.
	switch {
	case strings.HasPrefix(day, "do") || day == "domenica":
		return time.Sunday, nil
	case strings.HasPrefix(day, "lu") || day == "lunedì":
		return time.Monday, nil
	case strings.HasPrefix(day, "ma") || day == "martedì":
		return time.Tuesday, nil
	case strings.HasPrefix(day, "me") || day == "mercoledì":
		return time.Wednesday, nil
	case strings.HasPrefix(day, "gi") || day == "giovedì":
		return time.Thursday, nil
	case strings.HasPrefix(day, "ve") || day == "venerdì":
		return time.Friday, nil
	case strings.HasPrefix(day, "sa") || day == "sabato":
		return time.Saturday, nil
	default:
		// Java: throw new RuntimeException("Could not find day of week for '" + dayStr + "'")
		return 0, fmt.Errorf("could not find day of week for '%s'", dayStr)
	}
}

// GetMonth ports DateCheckFilter.getMonth(String) → time.Month (January=1).
// Italian prefixes gen…dic exactly as Java.
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(monthStr)
	switch {
	case strings.HasPrefix(mon, "gen"):
		return time.January, nil
	case strings.HasPrefix(mon, "feb"):
		return time.February, nil
	case strings.HasPrefix(mon, "mar"):
		return time.March, nil
	case strings.HasPrefix(mon, "apr"):
		return time.April, nil
	case strings.HasPrefix(mon, "mag"):
		return time.May, nil
	case strings.HasPrefix(mon, "giu"):
		return time.June, nil
	case strings.HasPrefix(mon, "lug"):
		return time.July, nil
	case strings.HasPrefix(mon, "ago"):
		return time.August, nil
	case strings.HasPrefix(mon, "set"):
		return time.September, nil
	case strings.HasPrefix(mon, "ott"):
		return time.October, nil
	case strings.HasPrefix(mon, "nov"):
		return time.November, nil
	case strings.HasPrefix(mon, "dic"):
		return time.December, nil
	default:
		// Java: throw new RuntimeException("Could not find month '" + monthStr + "'")
		return 0, fmt.Errorf("could not find month '%s'", monthStr)
	}
}
