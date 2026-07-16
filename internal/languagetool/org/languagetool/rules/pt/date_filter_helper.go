package pt

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// DateFilterHelper ports org.languagetool.rules.pt.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(trimSpecialPT(dayStr))
	switch {
	case strings.HasPrefix(day, "dom"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "seg"):
		return time.Monday, nil
	case strings.HasPrefix(day, "ter"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "qua"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "qui"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "sex"):
		return time.Friday, nil
	case strings.HasPrefix(day, "sáb"), strings.HasPrefix(day, "sab"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(trimSpecialPT(monthStr))
	switch {
	case strings.HasPrefix(mon, "jan"):
		return time.January, nil
	case strings.HasPrefix(mon, "fev"):
		return time.February, nil
	case strings.HasPrefix(mon, "mar"):
		return time.March, nil
	case strings.HasPrefix(mon, "abr"):
		return time.April, nil
	case strings.HasPrefix(mon, "mai"):
		return time.May, nil
	case strings.HasPrefix(mon, "jun"):
		return time.June, nil
	case strings.HasPrefix(mon, "jul"):
		return time.July, nil
	case strings.HasPrefix(mon, "ago"):
		return time.August, nil
	case strings.HasPrefix(mon, "set"):
		return time.September, nil
	case strings.HasPrefix(mon, "out"):
		return time.October, nil
	case strings.HasPrefix(mon, "nov"):
		return time.November, nil
	case strings.HasPrefix(mon, "dez"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}

func trimSpecialPT(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		if r == '\u00AD' || r == '.' {
			return -1
		}
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
