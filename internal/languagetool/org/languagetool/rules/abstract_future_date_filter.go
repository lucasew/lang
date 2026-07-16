package rules

import (
	"strconv"
	"strings"
	"time"
	"unicode"
)

// FutureDateFilterCore ports AbstractFutureDateFilter date assembly helpers.
type FutureDateFilterCore struct {
	// GetMonth maps a localized month name to 1–12; nil only accepts numeric months.
	GetMonth func(localized string) (int, error)
	// GetDayOfMonth maps a localized day name to 1–31; nil only accepts numeric days.
	GetDayOfMonth func(localized string) (int, error)
	// Now is the reference "today"; nil uses time.Now().UTC() (or test fixed date).
	Now func() time.Time
}

// ParseDay extracts a day-of-month from "22", "22nd", "22-an", etc.
func ParseDayOfMonthArg(s string, getNamed func(string) (int, error)) (int, error) {
	n := ""
	for _, r := range s {
		if unicode.IsDigit(r) {
			n += string(r)
		} else if n != "" {
			break
		}
	}
	if n != "" {
		return strconv.Atoi(n)
	}
	if getNamed != nil {
		return getNamed(s)
	}
	return 0, strconv.ErrSyntax
}

// ParseMonthArg accepts a numeric month or localized name via GetMonth.
func (f *FutureDateFilterCore) ParseMonthArg(monthStr string) (int, error) {
	monthStr = strings.TrimSpace(monthStr)
	if monthStr == "" {
		return 0, strconv.ErrSyntax
	}
	allDigit := true
	for _, r := range monthStr {
		if !unicode.IsDigit(r) {
			allDigit = false
			break
		}
	}
	if allDigit {
		return strconv.Atoi(monthStr)
	}
	if f.GetMonth != nil {
		return f.GetMonth(monthStr)
	}
	return 0, strconv.ErrSyntax
}

// IsFuture reports whether year/month/day (1-based month) is strictly after today.
func (f *FutureDateFilterCore) IsFuture(year, month, day int) bool {
	if year <= 0 || month < 1 || month > 12 || day < 1 {
		return false
	}
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	now := f.now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return d.After(today)
}

// AcceptFromArgs keeps the match when the date in args is in the future.
func (f *FutureDateFilterCore) AcceptFromArgs(args map[string]string) bool {
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
	// validate date exists
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	if t.Year() != y || int(t.Month()) != m || t.Day() != d {
		return false
	}
	return f.IsFuture(y, m, d)
}

func (f *FutureDateFilterCore) now() time.Time {
	if f.Now != nil {
		return f.Now()
	}
	if IsTest() {
		return time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	return time.Now().UTC()
}
