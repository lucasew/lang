package ca

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper ports org.languagetool.rules.ca.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch day {
	case "dg", "diumenge":
		return time.Sunday, nil
	case "dl", "dilluns":
		return time.Monday, nil
	case "dt", "dimarts":
		return time.Tuesday, nil
	case "dc", "dimecres":
		return time.Wednesday, nil
	case "dj", "dijous":
		return time.Thursday, nil
	case "dv", "divendres":
		return time.Friday, nil
	case "ds", "dissabte":
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
	case strings.HasPrefix(mon, "febr"):
		return time.February, nil
	case strings.HasPrefix(mon, "març"), strings.HasPrefix(mon, "marc"):
		return time.March, nil
	case strings.HasPrefix(mon, "abr"):
		return time.April, nil
	case strings.HasPrefix(mon, "maig"):
		return time.May, nil
	case strings.HasPrefix(mon, "juny"):
		return time.June, nil
	case strings.HasPrefix(mon, "jul"):
		return time.July, nil
	case strings.HasPrefix(mon, "ag"):
		return time.August, nil
	case strings.HasPrefix(mon, "set"):
		return time.September, nil
	case strings.HasPrefix(mon, "oct"):
		return time.October, nil
	case strings.HasPrefix(mon, "nov"):
		return time.November, nil
	case strings.HasPrefix(mon, "des"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
