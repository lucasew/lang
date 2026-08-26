package rules

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NewYearDateFilterCore ports org.languagetool.rules.AbstractNewYearDateFilter.
// Language filters supply GetMonth (and optionally GetDayOfMonth / GetCalendar).
type NewYearDateFilterCore struct {
	// GetMonth maps a localized month name to 1–12 (Java abstract getMonth).
	// Nil only accepts numeric months via StringUtils.isNumeric.
	GetMonth func(localized string) (int, error)
	// GetDayOfMonth maps spelled-out day; Java default returns 0.
	// Nil → only numeric days via DAY_OF_MONTH_PATTERN ((\\d+).*).
	GetDayOfMonth func(localized string) (int, error)
	// GetCalendar ports abstract getCalendar(); used by isJanuary/getCurrentYear
	// when not Force* / TestHackHelper. Nil → time.Now() (JVM default calendar).
	GetCalendar func() time.Time
	// ForceJanuary / ForceYear override calendar (tests).
	// Java uses TestHackHelper.isJUnitTest() only (January=true, year=2014).
	ForceJanuary *bool
	ForceYear    *int
}

// isJanuary ports AbstractNewYearDateFilter.isJanuary.
// TestHackHelper → true; else getCalendar().MONTH == JANUARY.
func (f *NewYearDateFilterCore) isJanuary() bool {
	if f != nil && f.ForceJanuary != nil {
		return *f.ForceJanuary
	}
	if IsTest() {
		return true
	}
	return f.calendarNow().Month() == time.January
}

// currentYear ports AbstractNewYearDateFilter.getCurrentYear.
// TestHackHelper → 2014; else getCalendar().YEAR.
func (f *NewYearDateFilterCore) currentYear() int {
	if f != nil && f.ForceYear != nil {
		return *f.ForceYear
	}
	if IsTest() {
		return 2014
	}
	return f.calendarNow().Year()
}

func (f *NewYearDateFilterCore) calendarNow() time.Time {
	if f != nil && f.GetCalendar != nil {
		return f.GetCalendar()
	}
	return time.Now()
}

// ShouldFlag is true in January for non-December dates whose year is currentYear-1.
// month is 1-based (January=1); Java compares Calendar.MONTH (December=11) after month-1.
func (f *NewYearDateFilterCore) ShouldFlag(year, month int) bool {
	// Java: isJanuary() && monthFromText != 11 /*December*/ && yearFromText + 1 == currentYear
	if !f.isJanuary() || month == 12 {
		return false
	}
	return year+1 == f.currentYear()
}

// FormatMessage replaces {year} and {realYear} placeholders (Java acceptRuleMatch).
func (f *NewYearDateFilterCore) FormatMessage(message string, yearFromText int) string {
	msg := strings.ReplaceAll(message, "{year}", strconv.Itoa(yearFromText))
	msg = strings.ReplaceAll(msg, "{realYear}", strconv.Itoa(f.currentYear()))
	return msg
}

// AcceptFromArgs returns rewritten message when the new-year condition holds; "" suppresses.
// Ports AbstractNewYearDateFilter.acceptRuleMatch date assembly + flag decision.
// Required args: year, month, day (Java getRequired → panic if missing).
// Soft hyphen U+00AD stripped on day and month (Java). Invalid calendar (setLenient false) → "".
func (f *NewYearDateFilterCore) AcceptFromArgs(args map[string]string, message string) string {
	if args == nil {
		return ""
	}
	// Java getRequired — IllegalArgumentException "Missing key '…'"
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
		// Java Integer.parseInt → NumberFormatException (propagates); grammar always supplies digits.
		return ""
	}
	// Java getMonthFromArguments: soft hyphen then isNumeric / getMonth; return month-1 for Calendar.
	// Go keeps 1-based and maps at the time.Month boundary (time.Month(m) == Java field + 1).
	monthStr := strings.ReplaceAll(args["month"], "\u00AD", "")
	m, err := f.parseMonth(monthStr)
	if err != nil {
		return ""
	}
	// Java getDayOfMonthFromArguments: soft hyphen then DAY_OF_MONTH_PATTERN / getDayOfMonth.
	dayStr := strings.ReplaceAll(args["day"], "\u00AD", "")
	d, err := ParseDayOfMonthArg(dayStr, f.dayOfMonthHook())
	if err != nil {
		return ""
	}
	// Java: calendar.setLenient(false); set(year, month-1, day, 0, 0, 0);
	// get(MONTH)/get(YEAR) throws IllegalArgumentException on invalid → acceptRuleMatch null.
	// Day 0 (Java getDayOfMonth default) is invalid under non-lenient calendar.
	loc := f.calendarLocation()
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc)
	if t.Year() != y || int(t.Month()) != m || t.Day() != d {
		return ""
	}
	// monthFromText from Calendar.MONTH is 0-based; Java compares != 11 (December).
	// Go ShouldFlag uses 1-based month (12 == December).
	if !f.ShouldFlag(y, m) {
		return ""
	}
	return f.FormatMessage(message, y)
}

func (f *NewYearDateFilterCore) dayOfMonthHook() func(string) (int, error) {
	if f != nil && f.GetDayOfMonth != nil {
		return f.GetDayOfMonth
	}
	return nil
}

func (f *NewYearDateFilterCore) calendarLocation() *time.Location {
	now := f.calendarNow()
	if now.Location() != nil {
		return now.Location()
	}
	return time.Local
}

// parseMonth ports getMonthFromArguments numeric branch without the Calendar month-1 adjust:
// returns 1-based month. Java: StringUtils.isNumeric → parseInt; else getMonth(monthStr)
// (no trimSpecialCharacters — DateFilterHelper may trim inside getMonth).
func (f *NewYearDateFilterCore) parseMonth(monthStr string) (int, error) {
	// Apache Commons StringUtils.isNumeric: non-empty, every char Character.isDigit
	if isAllDigits(monthStr) {
		return strconv.Atoi(monthStr)
	}
	if f != nil && f.GetMonth != nil {
		return f.GetMonth(monthStr)
	}
	return 0, fmt.Errorf("non-numeric month without GetMonth")
}
