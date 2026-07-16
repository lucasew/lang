package de

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// DateFilterHelper ports org.languagetool.rules.de.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetDayOfWeek returns Go time.Weekday for a German day name prefix.
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(trimSpecial(dayStr))
	switch {
	case strings.HasPrefix(day, "sonnabend"):
		return time.Saturday, nil
	case strings.HasPrefix(day, "so"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "mo"):
		return time.Monday, nil
	case strings.HasPrefix(day, "di"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "mi"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "do"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "fr"):
		return time.Friday, nil
	case strings.HasPrefix(day, "sa"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

// GetMonth returns month number 1–12 for a German month name prefix.
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(trimSpecial(monthStr))
	switch {
	case strings.HasPrefix(mon, "jän"), strings.HasPrefix(mon, "jan"):
		return time.January, nil
	case strings.HasPrefix(mon, "feb"):
		return time.February, nil
	case strings.HasPrefix(mon, "mär"):
		return time.March, nil
	case strings.HasPrefix(mon, "apr"):
		return time.April, nil
	case strings.HasPrefix(mon, "mai"):
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
	case strings.HasPrefix(mon, "dez"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}

func trimSpecial(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		// keep soft hyphen stripped etc.
		if r == '\u00AD' {
			return -1
		}
		if unicode.IsSpace(r) {
			return -1
		}
		// keep letters only for name matching
		if r == '.' {
			return -1
		}
		return r
	}, s)
}
