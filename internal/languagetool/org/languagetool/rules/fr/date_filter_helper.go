package fr

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper ports org.languagetool.rules.fr.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasPrefix(day, "dim"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "lun"):
		return time.Monday, nil
	case strings.HasPrefix(day, "mar"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "mer"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "jeu"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "ven"):
		return time.Friday, nil
	case strings.HasPrefix(day, "sam"):
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
	case strings.HasPrefix(mon, "fév"), strings.HasPrefix(mon, "fev"):
		return time.February, nil
	case strings.HasPrefix(mon, "mar"):
		return time.March, nil
	case strings.HasPrefix(mon, "avr"):
		return time.April, nil
	case strings.HasPrefix(mon, "mai"):
		return time.May, nil
	case strings.HasPrefix(mon, "juin"):
		return time.June, nil
	case strings.HasPrefix(mon, "juil"):
		return time.July, nil
	case strings.HasPrefix(mon, "aou"), strings.HasPrefix(mon, "aoû"):
		return time.August, nil
	case strings.HasPrefix(mon, "sep"):
		return time.September, nil
	case strings.HasPrefix(mon, "oct"):
		return time.October, nil
	case strings.HasPrefix(mon, "nov"):
		return time.November, nil
	case strings.HasPrefix(mon, "déc"), strings.HasPrefix(mon, "dec"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
