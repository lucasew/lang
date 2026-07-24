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
				// Java throws RuntimeException on unknown weekday.
				panic(err)
			}
			return wd
		},
		FormatDayOfWeek: func(t time.Time) string {
			// Java maps Locale.UK LONG English → Breton.
			switch t.Weekday() {
			case time.Sunday:
				return "Sul"
			case time.Monday:
				return "Lun"
			case time.Tuesday:
				return "Meurzh"
			case time.Wednesday:
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

// DateFilterHelper ports Breton day/month localization from DateCheckFilter.java.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetDayOfWeek ports DateCheckFilter.getDayOfWeek(String).
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasSuffix(day, "sul"):
		return time.Sunday, nil
	case strings.HasSuffix(day, "lun"):
		return time.Monday, nil
	case strings.HasSuffix(day, "meurzh"):
		return time.Tuesday, nil
	// typographic apostrophe as in Java source (U+2019)
	case strings.HasSuffix(day, "merc’her"), strings.HasSuffix(day, "merc'her"):
		return time.Wednesday, nil
	case day == "yaou", day == "diriaou":
		return time.Thursday, nil
	case strings.HasSuffix(day, "gwener"):
		return time.Friday, nil
	case strings.HasSuffix(day, "sadorn"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

// GetMonth ports DateCheckFilter.getMonth.
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(monthStr)
	switch mon {
	case "genver":
		return time.January, nil
	case "c’hwevrer", "c'hwevrer":
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
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}

// GetDayOfMonth ports DateCheckFilter.getDayOfMonth (spelled Breton day numbers).
func (h *DateFilterHelper) GetDayOfMonth(dayStr string) int {
	if dayStr == "" {
		return 0
	}
	day := strings.ToLower(dayStr)
	// soft mutation t→d, p→b (Java only in getDayOfMonth)
	if day[0] == 't' {
		day = "d" + day[1:]
	}
	if day[0] == 'p' {
		day = "b" + day[1:]
	}
	if strings.HasSuffix(day, "vet") {
		day = day[:len(day)-3]
	}
	switch day {
	case "c’hentañ", "c'hentañ", "unan":
		return 1
	case "daou", "eil":
		return 2
	case "dri", "drede", "deir":
		return 3
	case "bevar":
		return 4
	case "bemp", "bem":
		return 5
	case "c’hwerc’h", "c'hwerc'h", "c’hwerc'h", "c'hwerc’h":
		return 6
	case "seizh":
		return 7
	case "eizh":
		return 8
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
	case "c’hwezek", "c'hwezek":
		return 16
	case "seitek":
		return 17
	case "driwec’h", "driwec'h":
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
