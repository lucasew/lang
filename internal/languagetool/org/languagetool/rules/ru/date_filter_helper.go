package ru

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper ports Russian DateCheckFilter day/month localization.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetDayOfWeek ports DateCheckFilter.getDayOfWeek(String).
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasPrefix(day, "пн"), day == "понедельник":
		return time.Monday, nil
	case strings.HasPrefix(day, "вт"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "ср"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "чт"), day == "четверг":
		return time.Thursday, nil
	case day == "пт", strings.HasPrefix(day, "пятниц"):
		return time.Friday, nil
	case strings.HasPrefix(day, "сб"), strings.HasPrefix(day, "суббот"):
		return time.Saturday, nil
	case strings.HasPrefix(day, "вс"), day == "воскресенье":
		return time.Sunday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

// GetMonth ports DateCheckFilter.getMonth.
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	// Roman numerals compared on original (Java monthStr.equals("I") …).
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
	case "январь", "января", "янв":
		return time.January, nil
	case "февраль", "февраля", "фев":
		return time.February, nil
	case "март", "марта", "мар":
		return time.March, nil
	case "апрель", "апреля", "апр":
		return time.April, nil
	case "май", "мая":
		return time.May, nil
	case "июнь", "июня", "ин":
		return time.June, nil
	case "июль", "июля", "ил":
		return time.July, nil
	case "август", "августа", "авг":
		return time.August, nil
	case "сентябрь", "сентября", "сен":
		return time.September, nil
	case "октябрь", "октября", "окт":
		return time.October, nil
	case "ноябрь", "ноября", "ноя":
		return time.November, nil
	case "декабрь", "декабря", "дек":
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
