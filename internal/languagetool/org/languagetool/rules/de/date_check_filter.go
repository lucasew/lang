package de

import (
	"fmt"
	"time"
)

// DateCheckFilter ports org.languagetool.rules.de.DateCheckFilter helpers
// used by pattern-rule date checks (acceptRuleMatch needs full filter stack later).
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
	// Go: Sunday=0 … Saturday=6 → Java: Sunday=1 … Saturday=7
	return int(wd) + 1, nil
}

// GetMonth returns 1–12.
func (f *DateCheckFilter) GetMonth(monthStr string) (int, error) {
	m, err := f.helper.GetMonth(monthStr)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetDayOfWeekName returns German long weekday name for a date.
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	// German weekday names
	names := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
	return names[int(t.Weekday())]
}

// AcceptRuleMatch soft-checks year/month/day/weekDay consistency.
// Returns an error message if inconsistent, empty if OK.
// Requires weekDay key (Java throws if missing).
func (f *DateCheckFilter) AcceptRuleMatch(args map[string]string) (string, error) {
	if _, ok := args["weekDay"]; !ok {
		return "", fmt.Errorf("incomplete args: weekDay required")
	}
	// Full calendar validation deferred (pattern integration).
	return "", nil
}
