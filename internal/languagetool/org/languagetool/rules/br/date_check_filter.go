package br

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
	// soft mutation t→d, p→b (Breton)
	if strings.HasPrefix(d, "t") {
		d = "d" + d[1:]
	} else if strings.HasPrefix(d, "p") {
		d = "b" + d[1:]
	}
	switch {
	case d == "sul" || d == "sunday":
		return time.Sunday, nil
	case d == "lun" || d == "monday":
		return time.Monday, nil
	case d == "meurzh" || d == "tuesday":
		return time.Tuesday, nil
	case strings.HasPrefix(d, "merc"), d == "wednesday":
		return time.Wednesday, nil
	case d == "yaou" || d == "thursday":
		return time.Thursday, nil
	case d == "gwener" || d == "friday":
		return time.Friday, nil
	case d == "sadorn" || d == "saturday":
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	m := strings.ToLower(monthStr)
	switch {
	case strings.HasPrefix(m, "genver") || m == "1":
		return time.January, nil
	case strings.Contains(m, "hwevrer"), strings.HasPrefix(m, "chwevrer"), m == "2":
		return time.February, nil
	case strings.HasPrefix(m, "meurzh") || m == "3":
		return time.March, nil
	case strings.HasPrefix(m, "ebrel") || m == "4":
		return time.April, nil
	case m == "mae" || m == "5":
		return time.May, nil
	case strings.HasPrefix(m, "mezheven") || m == "6":
		return time.June, nil
	case strings.HasPrefix(m, "gouere") || m == "7":
		return time.July, nil
	case strings.HasPrefix(m, "eost") || m == "8":
		return time.August, nil
	case strings.HasPrefix(m, "gwengolo") || m == "9":
		return time.September, nil
	case strings.HasPrefix(m, "here") || m == "10":
		return time.October, nil
	case m == "du" || m == "11":
		return time.November, nil
	case strings.HasPrefix(m, "kerzu") || m == "12":
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
