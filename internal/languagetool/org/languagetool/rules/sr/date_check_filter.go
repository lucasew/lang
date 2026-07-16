package sr

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
	case strings.HasPrefix(d, "nedel"):
		return time.Sunday, nil
	case strings.HasPrefix(d, "poned"):
		return time.Monday, nil
	case strings.HasPrefix(d, "utor"):
		return time.Tuesday, nil
	case strings.HasPrefix(d, "sred"):
		return time.Wednesday, nil
	case strings.HasPrefix(d, "četv"), strings.HasPrefix(d, "cetv"):
		return time.Thursday, nil
	case strings.HasPrefix(d, "pet"):
		return time.Friday, nil
	case strings.HasPrefix(d, "sub"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	m := strings.ToLower(monthStr)
	switch {
	case strings.HasPrefix(m, "jan"):
		return time.January, nil
	case strings.HasPrefix(m, "feb"):
		return time.February, nil
	case strings.HasPrefix(m, "mar"):
		return time.March, nil
	case strings.HasPrefix(m, "apr"):
		return time.April, nil
	case strings.HasPrefix(m, "maj"):
		return time.May, nil
	case strings.HasPrefix(m, "jun"):
		return time.June, nil
	case strings.HasPrefix(m, "jul"):
		return time.July, nil
	case strings.HasPrefix(m, "avg"):
		return time.August, nil
	case strings.HasPrefix(m, "sep"):
		return time.September, nil
	case strings.HasPrefix(m, "okt"):
		return time.October, nil
	case strings.HasPrefix(m, "nov"):
		return time.November, nil
	case strings.HasPrefix(m, "dec"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
