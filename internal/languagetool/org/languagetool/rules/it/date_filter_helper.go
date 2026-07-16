package it

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper ports Italian DateCheckFilter day/month localization.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
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
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

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
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
