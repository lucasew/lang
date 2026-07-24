package rules

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NewYearDateFilterCore ports AbstractNewYearDateFilter acceptance logic.
type NewYearDateFilterCore struct {
	GetMonth func(localized string) (int, error)
	// GetDayOfMonth maps spelled-out day; Java default 0. Nil → only numeric days via pattern.
	GetDayOfMonth func(localized string) (int, error)
	// ForceJanuary / ForceYear override calendar (tests / Force* on language filters).
	ForceJanuary *bool
	ForceYear    *int
}

func (f *NewYearDateFilterCore) isJanuary() bool {
	if f.ForceJanuary != nil {
		return *f.ForceJanuary
	}
	if IsTest() {
		return true
	}
	return time.Now().Month() == time.January
}

func (f *NewYearDateFilterCore) currentYear() int {
	if f.ForceYear != nil {
		return *f.ForceYear
	}
	if IsTest() {
		return 2014
	}
	return time.Now().Year()
}

// ShouldFlag is true in January for non-December dates whose year is currentYear-1.
// month is 1-based (January=1); Java uses Calendar.MONTH (December=11).
func (f *NewYearDateFilterCore) ShouldFlag(year, month int) bool {
	if !f.isJanuary() || month == 12 {
		return false
	}
	return year+1 == f.currentYear()
}

// FormatMessage replaces {year} and {realYear} placeholders.
func (f *NewYearDateFilterCore) FormatMessage(message string, yearFromText int) string {
	msg := strings.ReplaceAll(message, "{year}", strconv.Itoa(yearFromText))
	msg = strings.ReplaceAll(msg, "{realYear}", strconv.Itoa(f.currentYear()))
	return msg
}

// AcceptFromArgs returns rewritten message when the new-year condition holds; "" suppresses.
// Java always builds a full Calendar (year/month/day) with setLenient(false); invalid day → null.
// Required args: year, month, day.
func (f *NewYearDateFilterCore) AcceptFromArgs(args map[string]string, message string) string {
	if args == nil {
		return ""
	}
	if _, ok := args["year"]; !ok {
		panic("Missing key 'year'")
	}
	if _, ok := args["month"]; !ok {
		panic("Missing key 'month'")
	}
	if _, ok := args["day"]; !ok {
		panic("Missing key 'day'")
	}
	y, err := strconv.Atoi(args["year"])
	if err != nil {
		return ""
	}
	m, err := f.parseMonth(args["month"])
	if err != nil {
		return ""
	}
	// Java: getRequired("day").replace soft hyphen before DAY_OF_MONTH_PATTERN
	dayStr := strings.ReplaceAll(args["day"], "\u00AD", "")
	d, err := ParseDayOfMonthArg(dayStr, f.GetDayOfMonth)
	if err != nil || d < 1 {
		return ""
	}
	// strict validity (Java setLenient(false) before get MONTH/YEAR)
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	if t.Year() != y || int(t.Month()) != m || t.Day() != d {
		return ""
	}
	if !f.ShouldFlag(y, m) {
		return ""
	}
	return f.FormatMessage(message, y)
}

func (f *NewYearDateFilterCore) parseMonth(monthStr string) (int, error) {
	// Java: StringUtils.isNumeric → parseInt; else getMonth (no trimSpecialCharacters)
	if isAllDigits(monthStr) {
		return strconv.Atoi(monthStr)
	}
	if f.GetMonth != nil {
		return f.GetMonth(monthStr)
	}
	return 0, fmt.Errorf("non-numeric month without GetMonth")
}
