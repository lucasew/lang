package pl

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper ports Polish DateCheckFilter day/month localization.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasPrefix(day, "pon"):
		return time.Monday, nil
	case strings.HasPrefix(day, "wt"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "śr"), strings.HasPrefix(day, "sr"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "czw"):
		return time.Thursday, nil
	case day == "pt" || strings.HasPrefix(day, "piątk") || strings.HasPrefix(day, "piatk") || day == "piątek" || day == "piatek":
		return time.Friday, nil
	case strings.HasPrefix(day, "sob"):
		return time.Saturday, nil
	case strings.HasPrefix(day, "niedz"):
		return time.Sunday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	// Roman numerals and genitive month names
	switch monthStr {
	case "I":
		return time.January, nil
	case "II":
		return time.February, nil
	case "III":
		return time.March, nil
	case "IV":
		return time.April, nil
	case "V":
		return time.May, nil
	case "VI":
		return time.June, nil
	case "VII":
		return time.July, nil
	case "VIII":
		return time.August, nil
	case "IX":
		return time.September, nil
	case "X":
		return time.October, nil
	case "XI":
		return time.November, nil
	case "XII":
		return time.December, nil
	}
	mon := strings.ToLower(monthStr)
	switch mon {
	case "stycznia", "styczeń", "styczen":
		return time.January, nil
	case "lutego", "luty":
		return time.February, nil
	case "marca", "marzec":
		return time.March, nil
	case "kwietnia", "kwiecień", "kwiecien":
		return time.April, nil
	case "maja", "maj":
		return time.May, nil
	case "czerwca", "czerwiec":
		return time.June, nil
	case "lipca", "lipiec":
		return time.July, nil
	case "sierpnia", "sierpień", "sierpien":
		return time.August, nil
	case "września", "wrzesnia", "wrzesień", "wrzesien":
		return time.September, nil
	case "października", "pazdziernika", "październik", "pazdziernik":
		return time.October, nil
	case "listopada", "listopad":
		return time.November, nil
	case "grudnia", "grudzień", "grudzien":
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
