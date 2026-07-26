package sr

import (
	"fmt"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// DateCheckFilter ports org.languagetool.rules.sr.DateCheckFilter
// (extends AbstractDateCheckFilter).
//
// Day/month localization is implemented here exactly as DateCheckFilter.java —
// SR has no separate DateFilterHelper type in Java.
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
	patterns.GlobalRuleFilterCreator.Register("org.languagetool.rules.sr.DateCheckFilter", func() patterns.RuleFilter {
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
//
//	startsWith("по") || equals("понедељак") → MONDAY
//	startsWith("ут") → TUESDAY
//	startsWith("ср") → WEDNESDAY
//	startsWith("че") || equals("четвртак") → THURSDAY
//	startsWith("пе") || equals("петак") → FRIDAY
//	startsWith("су") || equals("субота") → SATURDAY
//	startsWith("не") || equals("недеља") → SUNDAY
func (f *DateCheckFilter) getDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasPrefix(day, "по") || day == "понедељак":
		return time.Monday, nil
	case strings.HasPrefix(day, "ут"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "ср"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "че") || day == "четвртак":
		return time.Thursday, nil
	case strings.HasPrefix(day, "пе") || day == "петак":
		return time.Friday, nil
	case strings.HasPrefix(day, "су") || day == "субота":
		return time.Saturday, nil
	case strings.HasPrefix(day, "не") || day == "недеља":
		return time.Sunday, nil
	default:
		// Java: throw new RuntimeException("Редни број дана у недељи за '" + dayStr + "' не постоји.");
		return 0, fmt.Errorf("Редни број дана у недељи за '%s' не постоји.", dayStr)
	}
}

// getDayOfWeekName ports DateCheckFilter.getDayOfWeek(Calendar) —
// date.getDisplayName(DAY_OF_WEEK, LONG, Locale.forLanguageTag("sr")).
func (f *DateCheckFilter) getDayOfWeekName(t time.Time) string {
	// Java Locale "sr" LONG weekday names (Cyrillic; verified via Calendar.getDisplayName).
	names := []string{"недеља", "понедељак", "уторак", "среда", "четвртак", "петак", "субота"}
	return names[int(t.Weekday())]
}

// getMonth ports DateCheckFilter.getMonth(String) → 1..12.
// Serbian names compared after mon=monthStr.toLowerCase(); Roman numerals
// use original monthStr.equals (case-sensitive I..XII).
func (f *DateCheckFilter) getMonth(monthStr string) (int, error) {
	mon := strings.ToLower(monthStr)
	switch {
	case mon == "јануар" || monthStr == "I" || mon == "јануара" || mon == "јан":
		return 1, nil
	case mon == "фебруар" || monthStr == "II" || mon == "фебруара" || mon == "феб":
		return 2, nil
	case mon == "март" || monthStr == "III" || mon == "марта" || mon == "мар":
		return 3, nil
	case mon == "април" || monthStr == "IV" || mon == "априла" || mon == "апр":
		return 4, nil
	case mon == "мај" || monthStr == "V" || mon == "маја":
		return 5, nil
	// Java has redundant mon.equals("јун") twice — same outcome.
	case mon == "јун" || monthStr == "VI" || mon == "јуна":
		return 6, nil
	// Java has redundant mon.equals("јул") twice — same outcome.
	case mon == "јул" || monthStr == "VII" || mon == "јула":
		return 7, nil
	case mon == "август" || monthStr == "VIII" || mon == "августа" || mon == "авг":
		return 8, nil
	case mon == "септембар" || monthStr == "IX" || mon == "септембра" || mon == "сеп":
		return 9, nil
	case mon == "октобар" || monthStr == "X" || mon == "октобра" || mon == "окт":
		return 10, nil
	case mon == "новембар" || monthStr == "XI" || mon == "новембра" || mon == "нов":
		return 11, nil
	case mon == "децембар" || monthStr == "XII" || mon == "децембра" || mon == "дец":
		return 12, nil
	default:
		// Java: throw new RuntimeException("Месец '" + monthStr + "' не постоји.");
		return 0, fmt.Errorf("Месец '%s' не постоји.", monthStr)
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

// GetDayOfWeekName returns Serbian LONG weekday name for a calendar date.
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return f.getDayOfWeekName(t)
}
