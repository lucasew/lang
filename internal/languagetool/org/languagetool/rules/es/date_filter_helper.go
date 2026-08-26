package es

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper ports org.languagetool.rules.es.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch day {
	case "do", "domingo":
		return time.Sunday, nil
	case "lu", "lunes":
		return time.Monday, nil
	case "ma", "martes":
		return time.Tuesday, nil
	case "mi", "miércoles":
		return time.Wednesday, nil
	case "ju", "jueves":
		return time.Thursday, nil
	case "vi", "viernes":
		return time.Friday, nil
	case "sa", "sábado":
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(monthStr)
	switch {
	case strings.HasPrefix(mon, "en"):
		return time.January, nil
	case strings.HasPrefix(mon, "fe"):
		return time.February, nil
	case strings.HasPrefix(mon, "mar"), strings.HasPrefix(mon, "mzo"):
		return time.March, nil
	case strings.HasPrefix(mon, "ab"):
		return time.April, nil
	case strings.HasPrefix(mon, "may"), strings.HasPrefix(mon, "my"):
		return time.May, nil
	case strings.HasPrefix(mon, "jun"), mon == "jn":
		return time.June, nil
	case strings.HasPrefix(mon, "jul"), mon == "jl":
		return time.July, nil
	case strings.HasPrefix(mon, "ag"):
		return time.August, nil
	case strings.HasPrefix(mon, "se"):
		return time.September, nil
	case strings.HasPrefix(mon, "oc"):
		return time.October, nil
	case strings.HasPrefix(mon, "no"):
		return time.November, nil
	case strings.HasPrefix(mon, "di"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
