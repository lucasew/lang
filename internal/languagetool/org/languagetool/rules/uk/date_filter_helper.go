package uk

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper ports Ukrainian DateCheckFilter day/month localization.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasPrefix(day, "по") || day == "пн":
		return time.Monday, nil
	case strings.HasPrefix(day, "ві") || day == "вт":
		return time.Tuesday, nil
	case strings.HasPrefix(day, "се") || day == "ср":
		return time.Wednesday, nil
	case strings.HasPrefix(day, "че") || day == "чт":
		return time.Thursday, nil
	case strings.HasPrefix(day, "п'") || strings.HasPrefix(day, "п’") || day == "пт":
		return time.Friday, nil
	case strings.HasPrefix(day, "су") || day == "сб":
		return time.Saturday, nil
	case strings.HasPrefix(day, "не") || day == "нд":
		return time.Sunday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(monthStr)
	// note: "ли" is ambiguous (липень vs листопад) — Java has same order (липень first)
	switch {
	case strings.HasPrefix(mon, "сі"):
		return time.January, nil
	case strings.HasPrefix(mon, "лю"):
		return time.February, nil
	case strings.HasPrefix(mon, "бе"):
		return time.March, nil
	case strings.HasPrefix(mon, "кв"):
		return time.April, nil
	case strings.HasPrefix(mon, "тр"):
		return time.May, nil
	case strings.HasPrefix(mon, "че"):
		return time.June, nil
	case strings.HasPrefix(mon, "лип"):
		return time.July, nil
	case strings.HasPrefix(mon, "сер"):
		return time.August, nil
	case strings.HasPrefix(mon, "вер"):
		return time.September, nil
	case strings.HasPrefix(mon, "жов"):
		return time.October, nil
	case strings.HasPrefix(mon, "лис"):
		return time.November, nil
	case strings.HasPrefix(mon, "гр"):
		return time.December, nil
	case strings.HasPrefix(mon, "ли"):
		// bare "ли" → July per Java order
		return time.July, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
