package nl

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper ports day/month localization from Dutch DateCheckFilter.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch day {
	case "zo", "zondag":
		return time.Sunday, nil
	case "ma", "maandag":
		return time.Monday, nil
	case "di", "dinsdag":
		return time.Tuesday, nil
	case "wo", "woensdag":
		return time.Wednesday, nil
	case "do", "donderdag":
		return time.Thursday, nil
	case "vr", "vrijdag":
		return time.Friday, nil
	case "za", "zaterdag":
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(monthStr)
	switch {
	case strings.HasPrefix(mon, "jan"):
		return time.January, nil
	case strings.HasPrefix(mon, "feb"):
		return time.February, nil
	case strings.HasPrefix(mon, "mrt"), strings.HasPrefix(mon, "maa"):
		return time.March, nil
	case strings.HasPrefix(mon, "apr"):
		return time.April, nil
	case strings.HasPrefix(mon, "mei"):
		return time.May, nil
	case strings.HasPrefix(mon, "jun"):
		return time.June, nil
	case strings.HasPrefix(mon, "jul"):
		return time.July, nil
	case strings.HasPrefix(mon, "aug"):
		return time.August, nil
	case strings.HasPrefix(mon, "sep"):
		return time.September, nil
	case strings.HasPrefix(mon, "okt"):
		return time.October, nil
	case strings.HasPrefix(mon, "nov"):
		return time.November, nil
	case strings.HasPrefix(mon, "dec"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
