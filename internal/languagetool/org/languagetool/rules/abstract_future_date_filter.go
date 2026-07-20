package rules

import (
	"strconv"
	"time"
)

// FutureDateFilterCore ports AbstractFutureDateFilter date assembly helpers.
type FutureDateFilterCore struct {
	// GetMonth maps a localized month name to 1–12; nil only accepts numeric months.
	GetMonth func(localized string) (int, error)
	// GetDayOfMonth maps a localized day name to 1–31; nil only accepts numeric days.
	// Java getDayOfMonth default returns 0 (invalid → non-lenient calendar fails).
	GetDayOfMonth func(localized string) (int, error)
	// Now is the reference "today"; nil uses time.Now().UTC() (or test fixed date 2014-01-01).
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

// ParseMonthArg ports AbstractFutureDateFilter.getMonthFromArguments (1-based month).
// Java: StringUtils.isNumeric → parseInt; else getMonth(monthStr) (no trimSpecialCharacters).
func (f *FutureDateFilterCore) ParseMonthArg(monthStr string) (int, error) {
	// Apache Commons isNumeric: digits only
	if isAllDigits(monthStr) {
		return strconv.Atoi(monthStr)
	}
	if f != nil && f.GetMonth != nil {
		return f.GetMonth(monthStr)
	}
	return 0, strconv.ErrSyntax
}

// IsFuture reports whether year/month/day (1-based month) is strictly after "today"
// (Java dateFromDate.after(currentDate); test clock 2014-01-01).
func (f *FutureDateFilterCore) IsFuture(year, month, day int) bool {
	if year <= 0 || month < 1 || month > 12 || day < 1 {
		return false
	}
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	// strict validity
	if d.Year() != year || int(d.Month()) != month || d.Day() != day {
		return false
	}
	now := f.now()
	// Java compares full Calendar after(currentDate); times zeroed at midnight
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return d.After(today)
}

// AcceptFromArgs keeps the match when the date in args is in the future.
// Required args: year, month, day (Java getRequired).
func (f *FutureDateFilterCore) AcceptFromArgs(args map[string]string) bool {
	if args == nil {
		return false
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
		return false
	}
	m, err := f.ParseMonthArg(args["month"])
	if err != nil {
		return false
	}
	d, err := ParseDayOfMonthArg(args["day"], f.GetDayOfMonth)
	if err != nil {
		return false
	}
	// Java setLenient(false); invalid day (0) or Feb 30 → after() throws → null
	if d < 1 {
		return false
	}
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	if t.Year() != y || int(t.Month()) != m || t.Day() != d {
		return false
	}
	return f.IsFuture(y, m, d)
}

func (f *FutureDateFilterCore) now() time.Time {
	if f != nil && f.Now != nil {
		return f.Now()
	}
	if IsTest() {
		// Java TestHackHelper: currentDate = 2014-01-01
		return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	return time.Now().UTC()
}
