package uk

import (
	"fmt"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// DateCheckFilter ports org.languagetool.rules.uk.DateCheckFilter
// (extends AbstractDateCheckFilter).
//
// Day/month localization is implemented here exactly as DateCheckFilter.java —
// UK has no separate DateFilterHelper type in Java.
type DateCheckFilter struct {
	*rules.AbstractDateCheckFilter
}

func NewDateCheckFilter() *DateCheckFilter {
	f := &DateCheckFilter{}
	abs := &rules.AbstractDateCheckFilter{
		GetDayOfWeekName: func(localized string) time.Weekday {
			wd, err := f.getDayOfWeek(localized)
			if err != nil {
				// Java throws RuntimeException on unknown weekday.
				panic(err)
			}
			return wd
		},
		FormatDayOfWeek: func(t time.Time) string {
			return f.getDayOfWeekName(t)
		},
		GetMonth: func(localized string) int {
			m, err := f.getMonth(localized)
			if err != nil {
				// Java throws RuntimeException on unknown month.
				panic(err)
			}
			return m
		},
	}
	f.AbstractDateCheckFilter = abs
	return f
}

func init() {
	patterns.GlobalRuleFilterCreator.Register("org.languagetool.rules.uk.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch via AbstractDateCheckFilter.
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractDateCheckFilter == nil {
		return nil
	}
	return f.AbstractDateCheckFilter.AcceptRuleMatch(match, arguments)
}

// getDayOfWeek ports DateCheckFilter.getDayOfWeek(String).
// Only toLowerCase (no trimSpecialCharacters).
//
//	startsWith("по") || equals("пн") → MONDAY
//	startsWith("ві") || equals("вт") → TUESDAY
//	startsWith("се") || equals("ср") → WEDNESDAY
//	startsWith("че") || equals("чт") → THURSDAY
//	startsWith("п'") (U+0027) || startsWith("п'") (U+2019) || equals("пт") → FRIDAY
//	startsWith("су") || equals("сб") → SATURDAY
//	startsWith("не") || equals("нд") → SUNDAY
func (f *DateCheckFilter) getDayOfWeek(dayStr string) (time.Weekday, error) {
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
	// Java: startsWith("п'") || startsWith("п'") — U+0027 and U+2019
	case strings.HasPrefix(day, "п'") || strings.HasPrefix(day, "п\u2019") || day == "пт":
		return time.Friday, nil
	case strings.HasPrefix(day, "су") || day == "сб":
		return time.Saturday, nil
	case strings.HasPrefix(day, "не") || day == "нд":
		return time.Sunday, nil
	default:
		// Java: throw new RuntimeException("Could not find day of week for '" + dayStr + "'");
		return 0, fmt.Errorf("Could not find day of week for '%s'", dayStr)
	}
}

// getDayOfWeekName ports DateCheckFilter.getDayOfWeek(Calendar) —
// date.getDisplayName(DAY_OF_WEEK, LONG, Locale.forLanguageTag("uk")).
func (f *DateCheckFilter) getDayOfWeekName(t time.Time) string {
	// Java Locale "uk" LONG weekday names (nominative).
	// Friday uses U+02BC MODIFIER LETTER APOSTROPHE as in CLDR/uk for пʼятниця.
	names := []string{"неділя", "понеділок", "вівторок", "середа", "четвер", "пʼятниця", "субота"}
	return names[int(t.Weekday())]
}

// getMonth ports DateCheckFilter.getMonth(String) → 1..12 (Calendar.MONTH + 1).
// Java order includes a dead second startsWith("ли") for November — July wins first.
func (f *DateCheckFilter) getMonth(monthStr string) (int, error) {
	mon := strings.ToLower(monthStr)
	switch {
	case strings.HasPrefix(mon, "сі"):
		return 1, nil // Calendar.JANUARY + 1
	case strings.HasPrefix(mon, "лю"):
		return 2, nil
	case strings.HasPrefix(mon, "бе"):
		return 3, nil
	case strings.HasPrefix(mon, "кв"):
		return 4, nil
	case strings.HasPrefix(mon, "тр"):
		return 5, nil
	case strings.HasPrefix(mon, "че"):
		return 6, nil
	case strings.HasPrefix(mon, "ли"):
		// Java: July; second startsWith("ли") for November is unreachable dead code.
		return 7, nil
	case strings.HasPrefix(mon, "се"):
		return 8, nil
	case strings.HasPrefix(mon, "ве"):
		return 9, nil
	case strings.HasPrefix(mon, "жо"):
		return 10, nil
	// Java dead code: if (mon.startsWith("ли")) return Calendar.NOVEMBER + 1;
	// Kept as comment only — unreachable after the July branch above.
	case strings.HasPrefix(mon, "гр"):
		return 12, nil
	default:
		// Java: throw new RuntimeException("Could not find month '" + monthStr + "'");
		return 0, fmt.Errorf("Could not find month '%s'", monthStr)
	}
}

// GetDayOfWeekJava returns Java Calendar day-of-week (Sunday=1 … Saturday=7).
func (f *DateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	if f == nil {
		return 0, fmt.Errorf("nil DateCheckFilter")
	}
	wd, err := f.getDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	// Go Sunday=0 … Saturday=6 → Java Sunday=1 … Saturday=7
	return int(wd) + 1, nil
}

// GetMonth ports DateCheckFilter.getMonth for tests (1=January … 12=December).
func (f *DateCheckFilter) GetMonth(monthStr string) (int, error) {
	if f == nil {
		return 0, fmt.Errorf("nil DateCheckFilter")
	}
	return f.getMonth(monthStr)
}

// GetDayOfWeekName returns Ukrainian LONG weekday name for a calendar date.
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return f.getDayOfWeekName(t)
}
