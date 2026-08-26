package br

import (
	"fmt"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DateCheckFilter ports org.languagetool.rules.br.DateCheckFilter
// (extends AbstractDateCheckFilter).
type DateCheckFilter struct {
	*rules.AbstractDateCheckFilter
	helper *DateFilterHelper
}

func NewDateCheckFilter() *DateCheckFilter {
	h := NewDateFilterHelper()
	abs := &rules.AbstractDateCheckFilter{
		GetDayOfWeekName: func(localized string) time.Weekday {
			wd, err := h.GetDayOfWeek(localized)
			if err != nil {
				// Java throws RuntimeException on unknown weekday → AbstractDateCheckFilter
				// catches RuntimeException (not IllegalArgumentException) and returns null.
				panic(err)
			}
			return wd
		},
		// Java DateCheckFilter.getDayOfWeek(Calendar):
		// englishDay = date.getDisplayName(DAY_OF_WEEK, LONG, Locale.UK) then map to Breton.
		FormatDayOfWeek: func(t time.Time) string {
			return brDayOfWeekFromUKLong(t)
		},
		GetMonth: func(localized string) int {
			m, err := h.GetMonth(localized)
			if err != nil {
				// Java throws RuntimeException on unknown month.
				panic(err)
			}
			return int(m)
		},
		GetDayOfMonthOptional: h.GetDayOfMonth,
	}
	return &DateCheckFilter{AbstractDateCheckFilter: abs, helper: h}
}

// brDayOfWeekFromUKLong ports DateCheckFilter.getDayOfWeek(Calendar):
// Locale.UK LONG English → Breton names.
func brDayOfWeekFromUKLong(t time.Time) string {
	switch t.Weekday() {
	case time.Sunday:
		return "Sul"
	case time.Monday:
		return "Lun"
	case time.Tuesday:
		return "Meurzh"
	case time.Wednesday:
		// typographic apostrophe U+2019 as in Java source
		return "Merc’her"
	case time.Thursday:
		return "Yaou"
	case time.Friday:
		return "Gwener"
	case time.Saturday:
		return "Sadorn"
	default:
		return ""
	}
}

// GetDayOfWeekJava returns Java Calendar day-of-week (Sunday=1 … Saturday=7).
// Ports DateCheckFilter.getDayOfWeek(String).
func (f *DateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	if f == nil || f.helper == nil {
		return 0, fmt.Errorf("nil DateCheckFilter")
	}
	wd, err := f.helper.GetDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	// Go Sunday=0 … Saturday=6 → Java Sunday=1 … Saturday=7
	return int(wd) + 1, nil
}

// GetMonth ports DateCheckFilter.getMonth(String) → 1..12 for tests.
func (f *DateCheckFilter) GetMonth(monthStr string) (int, error) {
	if f == nil || f.helper == nil {
		return 0, fmt.Errorf("nil DateCheckFilter")
	}
	m, err := f.helper.GetMonth(monthStr)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

// GetDayOfWeekName ports DateCheckFilter.getDayOfWeek(Calendar) for a civil date.
func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return brDayOfWeekFromUKLong(t)
}

// GetDayOfMonth ports DateCheckFilter.getDayOfMonth(String) for tests.
func (f *DateCheckFilter) GetDayOfMonth(dayStr string) int {
	if f == nil || f.helper == nil {
		return 0
	}
	return f.helper.GetDayOfMonth(dayStr)
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch via AbstractDateCheckFilter.
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractDateCheckFilter == nil {
		return nil
	}
	return f.AbstractDateCheckFilter.AcceptRuleMatch(match, arguments)
}

// DateFilterHelper ports Breton day/month/day-of-month localization from DateCheckFilter.java.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetDayOfWeek ports DateCheckFilter.getDayOfWeek(String).
// endsWith/equals order matches Java sequential checks.
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasSuffix(day, "sul"):
		return time.Sunday, nil
	case strings.HasSuffix(day, "lun"):
		return time.Monday, nil
	case strings.HasSuffix(day, "meurzh"):
		return time.Tuesday, nil
	// typographic apostrophe U+2019 as in Java (merc’her); no ASCII-apostrophe invent
	case strings.HasSuffix(day, "merc’her"):
		return time.Wednesday, nil
	case day == "yaou":
		return time.Thursday, nil
	case day == "diriaou":
		return time.Thursday, nil
	case strings.HasSuffix(day, "gwener"):
		return time.Friday, nil
	case strings.HasSuffix(day, "sadorn"):
		return time.Saturday, nil
	default:
		// Java: throw new RuntimeException("Could not find day of week for '" + dayStr + "'")
		return 0, fmt.Errorf("could not find day of week for '%s'", dayStr)
	}
}

// GetMonth ports DateCheckFilter.getMonth(String) → time.Month (January=1).
// Full equals map only; unknown throws (RuntimeException in Java).
// Apostrophe in c’hwevrer is U+2019 only (no ASCII invent).
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(monthStr)
	switch mon {
	case "genver":
		return time.January, nil
	case "c’hwevrer":
		return time.February, nil
	case "meurzh":
		return time.March, nil
	case "ebrel":
		return time.April, nil
	case "mae":
		return time.May, nil
	case "mezheven", "even":
		return time.June, nil
	case "gouere", "gouhere":
		return time.July, nil
	case "eost":
		return time.August, nil
	case "gwengolo":
		return time.September, nil
	case "here":
		return time.October, nil
	case "du":
		return time.November, nil
	case "kerzu":
		return time.December, nil
	default:
		// Java: throw new RuntimeException("Could not find month '" + monthStr + "'")
		return 0, fmt.Errorf("could not find month '%s'", monthStr)
	}
}

// GetDayOfMonth ports DateCheckFilter.getDayOfMonth (spelled Breton day numbers).
// Soft mutation t→d, p→b; strip trailing "vet"; full equals map; 0 if unknown.
// Apostrophe forms use U+2019 only (Java source); no ASCII invent.
func (h *DateFilterHelper) GetDayOfMonth(dayStr string) int {
	day := strings.ToLower(dayStr)
	// Java: day.charAt(0) — empty dayStr throws (IndexOutOfBounds); no soft empty→0.
	if day[0] == 't' {
		day = "d" + day[1:]
	}
	if day[0] == 'p' {
		day = "b" + day[1:]
	}
	if strings.HasSuffix(day, "vet") {
		// Removing final vet if any.
		day = day[:len(day)-3]
	}
	switch day {
	case "c’hentañ", "unan":
		return 1
	case "daou", "eil":
		return 2
	case "dri", "drede", "deir":
		return 3
	case "bevar":
		return 4
	// bemp or bemvet (vet has been removed).
	case "bemp", "bem":
		return 5
	case "c’hwerc’h":
		return 6
	case "seizh":
		return 7
	case "eizh":
		return 8
	// nav and navet (vet has been removed).
	case "nav", "na":
		return 9
	case "dek":
		return 10
	case "unnek":
		return 11
	case "daouzek":
		return 12
	case "drizek":
		return 13
	case "bevarzek":
		return 14
	case "bemzek":
		return 15
	case "c’hwezek":
		return 16
	case "seitek":
		return 17
	case "driwec’h":
		return 18
	case "naontek":
		return 19
	case "ugent":
		return 20
	case "dregont":
		return 30
	default:
		return 0
	}
}
