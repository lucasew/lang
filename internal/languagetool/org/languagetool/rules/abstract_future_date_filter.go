package rules

import (
	"strconv"
	"time"
)

// FutureDateFilterCore ports org.languagetool.rules.AbstractFutureDateFilter.
// Language filters supply GetMonth (and optionally GetDayOfMonth); getCalendar/now
// is TestHackHelper-aware (fixed 2014-01-01 under tests).
type FutureDateFilterCore struct {
	// GetMonth maps a localized month name to 1–12 (Java abstract getMonth).
	// Nil only accepts numeric months via StringUtils.isNumeric.
	GetMonth func(localized string) (int, error)
	// GetDayOfMonth maps a localized day name to 1–31; Java default returns 0
	// (invalid → non-lenient calendar → after() throws → null).
	GetDayOfMonth func(localized string) (int, error)
	// Now is the reference "today" (Java getCalendar()); nil uses wall clock
	// (or TestHackHelper fixed date 2014-01-01 under go test).
	Now func() time.Time
}

// ParseDayOfMonthArg ports AbstractFutureDateFilter.getDayOfMonthFromArguments:
// DAY_OF_MONTH_PATTERN = (\d+).* with Matcher.matches(); else getDayOfMonth(name).
// Do not invent leading-digit scan (differs for e.g. partial mid-string digits).
func ParseDayOfMonthArg(s string, getNamed func(string) (int, error)) (int, error) {
	// Reuse full-region pattern from AbstractDateCheckFilter (same Java regex).
	if m := DayOfMonthPattern.FindStringSubmatch(s); m != nil {
		return strconv.Atoi(m[1])
	}
	if getNamed != nil {
		return getNamed(s)
	}
	// Java getDayOfMonth default → 0
	return 0, nil
}

// ParseMonthArg ports AbstractFutureDateFilter.getMonthFromArguments as 1-based month.
// Java returns month-1 for Calendar; callers using Go time.Month keep 1-based and
// map at the Calendar/time boundary (time.Month(m) == Java month field + 1).
// Java: StringUtils.isNumeric → parseInt; else getMonth(monthStr) (no trimSpecialCharacters).
func (f *FutureDateFilterCore) ParseMonthArg(monthStr string) (int, error) {
	// Apache Commons StringUtils.isNumeric: non-empty, every char Character.isDigit
	if isAllDigits(monthStr) {
		return strconv.Atoi(monthStr)
	}
	if f != nil && f.GetMonth != nil {
		return f.GetMonth(monthStr)
	}
	return 0, strconv.ErrSyntax
}

// getDate ports AbstractFutureDateFilter.getDate (year/month/day assembly,
// setLenient(false)). Returns ok=false for invalid calendar dates (Java after()
// throws IllegalArgumentException → acceptRuleMatch returns null).
// month/day args: month may be numeric or localized; day uses DAY_OF_MONTH_PATTERN.
func (f *FutureDateFilterCore) getDate(args map[string]string) (year, month, day int, ok bool) {
	y, err := strconv.Atoi(args["year"])
	if err != nil {
		// Java Integer.parseInt → NumberFormatException (propagates from getDate);
		// Go maps to soft drop (same as invalid calendar) — grammar always supplies digits.
		return 0, 0, 0, false
	}
	m, err := f.ParseMonthArg(args["month"])
	if err != nil {
		return 0, 0, 0, false
	}
	d, err := ParseDayOfMonthArg(args["day"], f.dayOfMonthHook())
	if err != nil {
		return 0, 0, 0, false
	}
	// Java: calendar.setLenient(false); set(year, month-1, day, 0, 0, 0)
	// Strict validity: reject rollover (e.g. Feb 30, day 0, month 13).
	loc := f.calendarLocation()
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc)
	if t.Year() != y || int(t.Month()) != m || t.Day() != d {
		return 0, 0, 0, false
	}
	return y, m, d, true
}

func (f *FutureDateFilterCore) dayOfMonthHook() func(string) (int, error) {
	if f != nil && f.GetDayOfMonth != nil {
		return f.GetDayOfMonth
	}
	return nil
}

// calendarLocation is the TZ used for date assembly and "today"
// (Java Calendar.getInstance(Locale.…) uses the JVM default timezone).
func (f *FutureDateFilterCore) calendarLocation() *time.Location {
	now := f.now()
	if now.Location() != nil {
		return now.Location()
	}
	return time.Local
}

// IsFuture reports whether year/month/day (1-based month) is strictly after "today"
// (Java dateFromDate.after(currentDate); test clock 2014-01-01 via TestHackHelper).
func (f *FutureDateFilterCore) IsFuture(year, month, day int) bool {
	loc := f.calendarLocation()
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
	// strict validity (Java setLenient(false); after() throws → not kept)
	if d.Year() != year || int(d.Month()) != month || d.Day() != day {
		return false
	}
	now := f.now()
	// Compare calendar dates at midnight in the same location.
	// (Java sets dateFromDate to 0:0:0; currentDate is getCalendar() / test fixed date.)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return d.After(today)
}

// AcceptFromArgs ports AbstractFutureDateFilter.acceptRuleMatch boolean decision:
// keep the match only when the date in args is strictly after currentDate.
// Required args: year, month, day (Java getRequired → panic if missing).
func (f *FutureDateFilterCore) AcceptFromArgs(args map[string]string) bool {
	if args == nil {
		// No Java twin for nil map; fail closed.
		return false
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
	y, m, d, ok := f.getDate(args)
	if !ok {
		// invalid calendar (lenient false) → after() IllegalArgumentException → null
		return false
	}
	return f.IsFuture(y, m, d)
}

// now ports getCalendar() + TestHackHelper.isJUnitTest() branch:
// under tests, currentDate = Calendar.Builder().setDate(2014, 0, 1).
func (f *FutureDateFilterCore) now() time.Time {
	if f != nil && f.Now != nil {
		return f.Now()
	}
	if IsTest() {
		// Java: new Calendar.Builder().setDate(2014, 0, 1).build()
		return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.Local)
	}
	// Java getCalendar() → Calendar.getInstance(locale) in system default TZ
	return time.Now()
}
