package eo

import (
	"fmt"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DateCheckFilter ports org.languagetool.rules.eo.DateCheckFilter
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
		// englishDay = date.getDisplayName(DAY_OF_WEEK, LONG, Locale.UK) then map to EO.
		FormatDayOfWeek: func(t time.Time) string {
			return eoDayOfWeekFromUKLong(t)
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

// eoDayOfWeekFromUKLong ports DateCheckFilter.getDayOfWeek(Calendar):
// Locale.UK LONG English → Esperanto names.
func eoDayOfWeekFromUKLong(t time.Time) string {
	switch t.Weekday() {
	case time.Sunday:
		return "dimanĉo"
	case time.Monday:
		return "lundo"
	case time.Tuesday:
		return "mardo"
	case time.Wednesday:
		return "merkredo"
	case time.Thursday:
		return "jaŭdo"
	case time.Friday:
		return "vendredo"
	case time.Saturday:
		return "sabato"
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
	return eoDayOfWeekFromUKLong(t)
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

// DateFilterHelper ports Esperanto day/month/day-of-month localization.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetDayOfWeek ports DateCheckFilter.getDayOfWeek(String).
// Prefix order matches Java sequential startsWith checks.
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasPrefix(day, "dim"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "lun"):
		return time.Monday, nil
	case strings.HasPrefix(day, "mar"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "mer"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "ĵaŭ"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "jau"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "jhau"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "jxau"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "ven"):
		return time.Friday, nil
	case strings.HasPrefix(day, "sab"):
		return time.Saturday, nil
	default:
		// Java: throw new RuntimeException("Could not find day of week for '" + dayStr + "'")
		return 0, fmt.Errorf("could not find day of week for '%s'", dayStr)
	}
}

// GetMonth ports DateCheckFilter.getMonth(String) → time.Month (January=1).
// Java only accepts August via startsWith("aŭg") — not Latin "aug".
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(monthStr)
	switch {
	case strings.HasPrefix(mon, "jan"):
		return time.January, nil
	case strings.HasPrefix(mon, "feb"):
		return time.February, nil
	case strings.HasPrefix(mon, "mar"):
		return time.March, nil
	case strings.HasPrefix(mon, "apr"):
		return time.April, nil
	case strings.HasPrefix(mon, "maj"):
		return time.May, nil
	case strings.HasPrefix(mon, "jun"):
		return time.June, nil
	case strings.HasPrefix(mon, "jul"):
		return time.July, nil
	case strings.HasPrefix(mon, "aŭg"):
		return time.August, nil
	case strings.HasPrefix(mon, "sep"):
		return time.September, nil
	case strings.HasPrefix(mon, "okt"):
		return time.October, nil
	case strings.HasPrefix(mon, "nov"):
		return time.November, nil
	case strings.HasPrefix(mon, "dec"):
		return time.December, nil
	default:
		// Java: throw new RuntimeException("Could not find month '" + monthStr + "'")
		return 0, fmt.Errorf("could not find month '%s'", monthStr)
	}
}

// GetDayOfMonth ports DateCheckFilter.getDayOfMonth (spelled Esperanto ordinals).
// Control flow is bug-for-bug with Java: final -n strip; dek / else dudek / else tridek;
// optional hyphen; equals unua…oka / naŭa / nauxa / naua (separate ifs in Java).
func (h *DateFilterHelper) GetDayOfMonth(dayStr string) int {
	day := strings.ToLower(dayStr)
	if strings.HasSuffix(day, "n") {
		// Removing final n if any (accusative).
		day = day[:len(day)-1]
	}
	n := 0
	// Java order: if dek else if dudek else if tridek (dudek/tridek do not start with "dek").
	if strings.HasPrefix(day, "dek") {
		n = 10
		day = day[3:]
	} else if strings.HasPrefix(day, "dudek") {
		n = 20
		day = day[5:]
	} else if strings.HasPrefix(day, "tridek") {
		n = 30
		day = day[6:]
	}
	if n > 0 && strings.HasPrefix(day, "-") {
		// Remove hyphen as in "dudek-trian".
		day = day[1:]
	}
	// Java uses separate if equals (not if-else); mutually exclusive for these literals.
	if day == "unua" {
		n += 1
	}
	if day == "dua" {
		n += 2
	}
	if day == "tria" {
		n += 3
	}
	if day == "kvara" {
		n += 4
	}
	if day == "kvina" {
		n += 5
	}
	if day == "sesa" {
		n += 6
	}
	if day == "sepa" {
		n += 7
	}
	if day == "oka" {
		n += 8
	}
	if day == "naŭa" {
		n += 9
	}
	if day == "nauxa" {
		n += 9
	}
	if day == "naua" {
		n += 9
	}
	return n
}
