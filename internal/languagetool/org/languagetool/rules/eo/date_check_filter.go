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
				panic(err)
			}
			return wd
		},
		FormatDayOfWeek: func(t time.Time) string {
			// Java maps Locale.UK LONG → Esperanto.
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
		},
		GetMonth: func(localized string) int {
			m, err := h.GetMonth(localized)
			if err != nil {
				return 0
			}
			return int(m)
		},
		GetDayOfMonthOptional: h.GetDayOfMonth,
	}
	return &DateCheckFilter{AbstractDateCheckFilter: abs, helper: h}
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
	case strings.HasPrefix(day, "ĵaŭ"), strings.HasPrefix(day, "jau"),
		strings.HasPrefix(day, "jhau"), strings.HasPrefix(day, "jxau"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "ven"):
		return time.Friday, nil
	case strings.HasPrefix(day, "sab"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

// GetMonth ports DateCheckFilter.getMonth.
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
	case strings.HasPrefix(mon, "aŭg"), strings.HasPrefix(mon, "aug"):
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
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}

// GetDayOfMonth ports DateCheckFilter.getDayOfMonth (spelled Esperanto ordinals).
func (h *DateFilterHelper) GetDayOfMonth(dayStr string) int {
	day := strings.ToLower(dayStr)
	if strings.HasSuffix(day, "n") {
		// Removing final n if any (accusative).
		day = day[:len(day)-1]
	}
	n := 0
	// Java order: dek, then dudek, then tridek.
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
	switch day {
	case "unua":
		n += 1
	case "dua":
		n += 2
	case "tria":
		n += 3
	case "kvara":
		n += 4
	case "kvina":
		n += 5
	case "sesa":
		n += 6
	case "sepa":
		n += 7
	case "oka":
		n += 8
	case "naŭa", "nauxa", "naua":
		n += 9
	}
	return n
}
