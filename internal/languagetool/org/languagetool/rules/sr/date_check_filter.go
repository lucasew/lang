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
// (extends AbstractDateCheckFilter with Serbian localization).
type DateCheckFilter struct {
	helper *DateFilterHelper
	inner  *rules.AbstractDateCheckFilter
}

func NewDateCheckFilter() *DateCheckFilter {
	return &DateCheckFilter{
		helper: NewDateFilterHelper(),
		inner:  srDateCheckFilter(),
	}
}

func srDateCheckFilter() *rules.AbstractDateCheckFilter {
	h := NewDateFilterHelper()
	return &rules.AbstractDateCheckFilter{
		GetDayOfWeekName: func(localized string) time.Weekday {
			wd, err := h.GetDayOfWeek(localized)
			if err != nil {
				panic(err)
			}
			return wd
		},
		FormatDayOfWeek: formatSRDayOfWeek,
		GetMonth: func(localized string) int {
			m, err := h.GetMonth(localized)
			if err != nil {
				return 0
			}
			return int(m)
		},
	}
}

func formatSRDayOfWeek(t time.Time) string {
	// Java: Calendar LONG display name for Locale "sr" (Cyrillic Serbian)
	names := []string{"недеља", "понедељак", "уторак", "среда", "четвртак", "петак", "субота"}
	return names[int(t.Weekday())]
}

func init() {
	patterns.GlobalRuleFilterCreator.Register("org.languagetool.rules.sr.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
}

func (f *DateCheckFilter) GetDayOfWeekJava(dayStr string) (int, error) {
	wd, err := f.helper.GetDayOfWeek(dayStr)
	if err != nil {
		return 0, err
	}
	return int(wd) + 1, nil
}

func (f *DateCheckFilter) GetMonth(monthStr string) (int, error) {
	m, err := f.helper.GetMonth(monthStr)
	if err != nil {
		return 0, err
	}
	return int(m), nil
}

func (f *DateCheckFilter) GetDayOfWeekName(year, month, day int) string {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return formatSRDayOfWeek(t)
}

// AcceptRuleMatch ports DateCheckFilter.acceptRuleMatch (super AbstractDateCheckFilter).
func (f *DateCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if f.inner == nil {
		f.inner = srDateCheckFilter()
	}
	return f.inner.AcceptRuleMatch(match, arguments)
}

// DateFilterHelper ports Serbian DateCheckFilter day/month localization (Java twin).
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	// Java: toLowerCase on dayStr; Cyrillic prefixes
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
		// Java: RuntimeException with Serbian message
		return 0, fmt.Errorf("редни број дана у недељи за '%s' не постоји", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	// Java compares mon = toLowerCase and original monthStr for Roman numerals
	mon := strings.ToLower(monthStr)
	switch {
	case mon == "јануар" || monthStr == "I" || mon == "јануара" || mon == "јан":
		return time.January, nil
	case mon == "фебруар" || monthStr == "II" || mon == "фебруара" || mon == "феб":
		return time.February, nil
	case mon == "март" || monthStr == "III" || mon == "марта" || mon == "мар":
		return time.March, nil
	case mon == "април" || monthStr == "IV" || mon == "априла" || mon == "апр":
		return time.April, nil
	case mon == "мај" || monthStr == "V" || mon == "маја":
		return time.May, nil
	case mon == "јун" || monthStr == "VI" || mon == "јуна":
		return time.June, nil
	case mon == "јул" || monthStr == "VII" || mon == "јула":
		return time.July, nil
	case mon == "август" || monthStr == "VIII" || mon == "августа" || mon == "авг":
		return time.August, nil
	case mon == "септембар" || monthStr == "IX" || mon == "септембра" || mon == "сеп":
		return time.September, nil
	case mon == "октобар" || monthStr == "X" || mon == "октобра" || mon == "окт":
		return time.October, nil
	case mon == "новембар" || monthStr == "XI" || mon == "новембра" || mon == "нов":
		return time.November, nil
	case mon == "децембар" || monthStr == "XII" || mon == "децембра" || mon == "дец":
		return time.December, nil
	default:
		return 0, fmt.Errorf("месец '%s' не постоји", monthStr)
	}
}
