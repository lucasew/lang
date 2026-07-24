package rules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DayOfMonthPattern ports AbstractDateCheckFilter.DAY_OF_MONTH_PATTERN:
// Pattern.compile("(\\d+).*") used with Matcher.matches() (full region).
// Capture group 1 is the numeric day (e.g. "22" from "22nd").
var DayOfMonthPattern = regexp.MustCompile(`\A(\d+).*\z`)

// AbstractDateCheckFilter ports org.languagetool.rules.AbstractDateCheckFilter.
// Language modules supply weekday/month localization hooks.
type AbstractDateCheckFilter struct {
	// GetDayOfWeekName maps localized weekday string → time.Weekday (Sunday=0 in Go).
	// Java uses Calendar where Sunday=1; language hooks convert at the boundary.
	GetDayOfWeekName func(localized string) time.Weekday
	// FormatDayOfWeek returns localized weekday name for a date.
	FormatDayOfWeek func(t time.Time) string
	// GetMonth maps localized month → 1..12 (January=1).
	GetMonth func(localized string) int
	// GetDayOfMonthOptional maps spelled-out day-of-month; 0 if unknown.
	GetDayOfMonthOptional func(localized string) int
	// Now is the reference "today" (defaults to time.Now).
	Now func() time.Time
	// TestMode forces year 2014 when year arg missing (Java JUnit / TestHackHelper hack).
	TestMode bool
}

// AcceptRuleMatch keeps the match when claimed weekday != actual weekday for the date.
// Args: year (optional), month, day, weekDay.
//
// Java: IllegalArgumentException from getRequired is rethrown; other RuntimeException
// from localization mapping is logged and returns null.
func (f *AbstractDateCheckFilter) AcceptRuleMatch(match *RuleMatch, args map[string]string) *RuleMatch {
	if f == nil || f.GetDayOfWeekName == nil || f.GetMonth == nil || f.FormatDayOfWeek == nil {
		return nil
	}
	// Required keys — Java getRequired → IllegalArgumentException (must panic, not recover).
	weekDayStr, ok := args["weekDay"]
	if !ok {
		panic("Missing key 'weekDay'")
	}
	if _, ok := args["day"]; !ok {
		panic("Missing key 'day'")
	}
	if _, ok := args["month"]; !ok {
		panic("Missing key 'month'")
	}

	// Java RuntimeException from localization mapping → skip match (do not crash).
	var out *RuleMatch
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Re-raise missing-key panics if any nested path still uses them.
				if s, ok := r.(string); ok && strings.HasPrefix(s, "Missing key ") {
					panic(r)
				}
				out = nil
			}
		}()
		out = f.acceptRuleMatchInner(match, args, weekDayStr)
	}()
	return out
}

func (f *AbstractDateCheckFilter) acceptRuleMatchInner(match *RuleMatch, args map[string]string, weekDayStr string) *RuleMatch {
	// Java: replace soft hyphen U+00AD
	weekDayStr = strings.ReplaceAll(weekDayStr, "\u00AD", "")
	claimed := f.GetDayOfWeekName(weekDayStr)
	date, year, err := f.parseDate(args)
	if err != nil {
		// invalid calendar date (e.g. 32.8.2014) — Java IllegalArgumentException on get DAY_OF_WEEK
		return nil
	}
	actual := date.Weekday()
	if claimed == actual {
		return nil
	}
	msg := match.GetMessage()
	msg = strings.ReplaceAll(msg, "{realDay}", f.FormatDayOfWeek(date))
	// Java: Calendar.set(DAY_OF_WEEK, claimed) then getDayOfWeek(cal)
	ref := time.Date(2014, 1, 5, 0, 0, 0, 0, time.UTC) // Sunday
	for ref.Weekday() != claimed {
		ref = ref.AddDate(0, 0, 1)
	}
	msg = strings.ReplaceAll(msg, "{day}", f.FormatDayOfWeek(ref))
	now := time.Now()
	if f.Now != nil {
		now = f.Now()
	}
	msg = strings.ReplaceAll(msg, "{currentYear}", strconv.Itoa(now.Year()))
	rm := NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	rm.ShortMessage = match.ShortMessage
	// Java: ruleMatch.setUrl(Tools.getUrl("https://www.timeanddate.com/calendar/?year=" + year))
	rm.SetURL(fmt.Sprintf("https://www.timeanddate.com/calendar/?year=%d", year))
	return rm
}

func (f *AbstractDateCheckFilter) parseDate(args map[string]string) (time.Time, int, error) {
	yearArg := args["year"]
	var year int
	now := time.Now()
	if f.Now != nil {
		now = f.Now()
	}
	if yearArg == "" {
		if f.TestMode {
			// Java TestHackHelper.isJUnitTest() → year 2014
			year = 2014
		} else {
			year = now.Year()
		}
	} else {
		y, err := strconv.Atoi(yearArg)
		if err != nil {
			return time.Time{}, 0, err
		}
		year = y
	}
	month, err := f.monthFromArgs(args)
	if err != nil {
		return time.Time{}, 0, err
	}
	day, err := f.dayFromArgs(args)
	if err != nil {
		return time.Time{}, 0, err
	}
	// Java calendar.setLenient(false) — strict validity
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if t.Year() != year || int(t.Month()) != month || t.Day() != day {
		return time.Time{}, 0, fmt.Errorf("invalid date")
	}
	return t, year, nil
}

func (f *AbstractDateCheckFilter) dayFromArgs(args map[string]string) (int, error) {
	dayStr, ok := args["day"]
	if !ok {
		panic("Missing key 'day'")
	}
	// Java Matcher.matches() on DAY_OF_MONTH_PATTERN — full string, not invent partial Find.
	if m := DayOfMonthPattern.FindStringSubmatch(dayStr); m != nil {
		return strconv.Atoi(m[1])
	}
	// Java getDayOfMonth (default 0); day 0 + non-lenient calendar fails later.
	if f.GetDayOfMonthOptional != nil {
		d := f.GetDayOfMonthOptional(dayStr)
		if d != 0 {
			return d, nil
		}
		return 0, fmt.Errorf("bad day: %s", dayStr)
	}
	return 0, fmt.Errorf("bad day: %s", dayStr)
}

func (f *AbstractDateCheckFilter) monthFromArgs(args map[string]string) (int, error) {
	monthStr, ok := args["month"]
	if !ok {
		panic("Missing key 'month'")
	}
	// Java: org.apache.commons.lang3.StringUtils.isNumeric — digits only, empty false.
	// (Not StringTools.isNumeric which allows spaces/commas.)
	if isAllDigits(monthStr) {
		m, err := strconv.Atoi(monthStr)
		return m, err // 1-based; Java later does month-1 for Calendar
	}
	// Java: getMonth(StringTools.trimSpecialCharacters(monthStr))
	return f.GetMonth(tools.TrimSpecialCharacters(monthStr)), nil
}
