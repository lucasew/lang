package rules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// DayOfMonthPattern extracts the numeric day from forms like "22nd".
var DayOfMonthPattern = regexp.MustCompile(`(\d+).*`)

// AbstractDateCheckFilter ports org.languagetool.rules.AbstractDateCheckFilter.
// Language modules supply weekday/month localization hooks.
type AbstractDateCheckFilter struct {
	// GetDayOfWeekName maps localized weekday string → time.Weekday (Sunday=0 in Go).
	// Java uses Calendar where Sunday=1; we convert at the comparison boundary.
	GetDayOfWeekName func(localized string) time.Weekday
	// FormatDayOfWeek returns localized weekday name for a date.
	FormatDayOfWeek func(t time.Time) string
	// GetMonth maps localized month → 1..12 (January=1).
	GetMonth func(localized string) int
	// GetDayOfMonthOptional maps spelled-out day-of-month; 0 if unknown.
	GetDayOfMonthOptional func(localized string) int
	// Now is the reference "today" (defaults to time.Now).
	Now func() time.Time
	// TestMode forces year 2014 when year arg missing (Java JUnit hack).
	TestMode bool
}

// AcceptRuleMatch keeps the match when claimed weekday != actual weekday for the date.
// Args: year (optional), month, day, weekDay.
func (f *AbstractDateCheckFilter) AcceptRuleMatch(match *RuleMatch, args map[string]string) *RuleMatch {
	if f == nil || f.GetDayOfWeekName == nil || f.GetMonth == nil || f.FormatDayOfWeek == nil {
		return nil
	}
	weekDayStr, ok := args["weekDay"]
	if !ok {
		panic("Missing key 'weekDay'")
	}
	weekDayStr = strings.ReplaceAll(weekDayStr, "\u00AD", "")
	claimed := f.GetDayOfWeekName(weekDayStr)
	date, err := f.parseDate(args)
	if err != nil {
		// invalid calendar date (e.g. 32.8.2014)
		return nil
	}
	actual := date.Weekday()
	if claimed == actual {
		return nil
	}
	// build message replacements
	msg := match.GetMessage()
	msg = strings.ReplaceAll(msg, "{realDay}", f.FormatDayOfWeek(date))
	// claimed weekday name via a calendar set to that weekday in a reference week
	ref := time.Date(2014, 1, 5, 0, 0, 0, 0, time.UTC) // Sunday 2014-01-05
	// move to claimed weekday
	for ref.Weekday() != claimed {
		ref = ref.AddDate(0, 0, 1)
	}
	msg = strings.ReplaceAll(msg, "{day}", f.FormatDayOfWeek(ref))
	now := time.Now()
	if f.Now != nil {
		now = f.Now()
	}
	msg = strings.ReplaceAll(msg, "{currentYear}", strconv.Itoa(now.Year()))
	out := NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	out.ShortMessage = match.ShortMessage
	return out
}

func (f *AbstractDateCheckFilter) parseDate(args map[string]string) (time.Time, error) {
	yearArg := args["year"]
	var year int
	now := time.Now()
	if f.Now != nil {
		now = f.Now()
	}
	if yearArg == "" {
		if f.TestMode {
			year = 2014
		} else {
			year = now.Year()
		}
	} else {
		y, err := strconv.Atoi(yearArg)
		if err != nil {
			return time.Time{}, err
		}
		year = y
	}
	month, err := f.monthFromArgs(args)
	if err != nil {
		return time.Time{}, err
	}
	day, err := f.dayFromArgs(args)
	if err != nil {
		return time.Time{}, err
	}
	// strict validity
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if t.Year() != year || int(t.Month()) != month || t.Day() != day {
		return time.Time{}, fmt.Errorf("invalid date")
	}
	return t, nil
}

func (f *AbstractDateCheckFilter) dayFromArgs(args map[string]string) (int, error) {
	dayStr, ok := args["day"]
	if !ok {
		panic("Missing key 'day'")
	}
	if m := DayOfMonthPattern.FindStringSubmatch(dayStr); m != nil {
		return strconv.Atoi(m[1])
	}
	if f.GetDayOfMonthOptional != nil {
		if d := f.GetDayOfMonthOptional(dayStr); d > 0 {
			return d, nil
		}
	}
	return 0, fmt.Errorf("bad day: %s", dayStr)
}

func (f *AbstractDateCheckFilter) monthFromArgs(args map[string]string) (int, error) {
	monthStr, ok := args["month"]
	if !ok {
		panic("Missing key 'month'")
	}
	if isAllDigits(monthStr) {
		m, err := strconv.Atoi(monthStr)
		return m, err // 1-based
	}
	return f.GetMonth(trimSpecialChars(monthStr)), nil
}

func trimSpecialChars(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}
