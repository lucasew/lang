package ru

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// DateFilterHelper ports org.languagetool.rules.ru.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(trimSpecialRU(dayStr))
	switch {
	case strings.HasPrefix(day, "суб"), strings.HasPrefix(day, "сб"):
		return time.Saturday, nil
	case strings.HasPrefix(day, "вс"), strings.HasPrefix(day, "вос"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "пн"), strings.HasPrefix(day, "пон"):
		return time.Monday, nil
	case strings.HasPrefix(day, "вт"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "ср"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "чт"), strings.HasPrefix(day, "чет"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "пт"), strings.HasPrefix(day, "пят"):
		return time.Friday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(trimSpecialRU(monthStr))
	switch {
	case strings.HasPrefix(mon, "янв"):
		return time.January, nil
	case strings.HasPrefix(mon, "фев"):
		return time.February, nil
	case strings.HasPrefix(mon, "мар"):
		return time.March, nil
	case strings.HasPrefix(mon, "апр"):
		return time.April, nil
	case strings.HasPrefix(mon, "май"), strings.HasPrefix(mon, "мая"):
		return time.May, nil
	case strings.HasPrefix(mon, "июн"):
		return time.June, nil
	case strings.HasPrefix(mon, "июл"):
		return time.July, nil
	case strings.HasPrefix(mon, "авг"):
		return time.August, nil
	case strings.HasPrefix(mon, "сен"):
		return time.September, nil
	case strings.HasPrefix(mon, "окт"):
		return time.October, nil
	case strings.HasPrefix(mon, "ноя"):
		return time.November, nil
	case strings.HasPrefix(mon, "дек"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}

func trimSpecialRU(s string) string {
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
