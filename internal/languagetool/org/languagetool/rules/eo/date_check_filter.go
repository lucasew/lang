package eo

import (
	"fmt"
	"strings"
	"time"
)

// DateCheckFilter ports language-local DateCheckFilter for pattern rules.
type DateCheckFilter struct {
	helper *DateFilterHelper
}

func NewDateCheckFilter() *DateCheckFilter {
	return &DateCheckFilter{helper: NewDateFilterHelper()}
}

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

func (f *DateCheckFilter) AcceptRuleMatch(args map[string]string) (string, error) {
	if _, ok := args["weekDay"]; !ok {
		return "", fmt.Errorf("incomplete args: weekDay required")
	}
	return "", nil
}

// DateFilterHelper localizes day/month names.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	d := strings.ToLower(dayStr)
	switch {
	case strings.HasPrefix(d, "diman"):
		return time.Sunday, nil
	case strings.HasPrefix(d, "lundo"):
		return time.Monday, nil
	case strings.HasPrefix(d, "mardo"):
		return time.Tuesday, nil
	case strings.HasPrefix(d, "merkredo"):
		return time.Wednesday, nil
	case strings.HasPrefix(d, "ĵaŭdo"), strings.HasPrefix(d, "jaudo"):
		return time.Thursday, nil
	case strings.HasPrefix(d, "vendredo"):
		return time.Friday, nil
	case strings.HasPrefix(d, "sabato"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	m := strings.ToLower(monthStr)
	switch {
	case strings.HasPrefix(m, "januar"):
		return time.January, nil
	case strings.HasPrefix(m, "februar"):
		return time.February, nil
	case strings.HasPrefix(m, "mart"):
		return time.March, nil
	case strings.HasPrefix(m, "april"):
		return time.April, nil
	case strings.HasPrefix(m, "maj"):
		return time.May, nil
	case strings.HasPrefix(m, "juni"):
		return time.June, nil
	case strings.HasPrefix(m, "juli"):
		return time.July, nil
	case strings.HasPrefix(m, "aŭgust"), strings.HasPrefix(m, "august"):
		return time.August, nil
	case strings.HasPrefix(m, "septembr"):
		return time.September, nil
	case strings.HasPrefix(m, "oktobr"):
		return time.October, nil
	case strings.HasPrefix(m, "novembr"):
		return time.November, nil
	case strings.HasPrefix(m, "decembr"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
