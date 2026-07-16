package nl

import (
	"fmt"
	"time"
)

// DateCheckFilter ports language-local DateCheckFilter helpers for pattern rules.
type DateCheckFilter struct {
	helper *DateFilterHelper
}

func NewDateCheckFilter() *DateCheckFilter {
	return &DateCheckFilter{helper: NewDateFilterHelper()}
}

// GetDayOfWeekJava returns Java Calendar day-of-week (Sunday=1 … Saturday=7).
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

func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	names := []string{"zondag", "maandag", "dinsdag", "woensdag", "donderdag", "vrijdag", "zaterdag"}
	return names[int(t.Weekday())]
}

func (f *DateCheckFilter) AcceptRuleMatch(args map[string]string) (string, error) {
	if _, ok := args["weekDay"]; !ok {
		return "", fmt.Errorf("incomplete args: weekDay required")
	}
	return "", nil
}
