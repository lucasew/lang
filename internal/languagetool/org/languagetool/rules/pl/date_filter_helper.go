package pl

import (
	"fmt"
	"strings"
	"time"
)

// DateFilterHelper holds Polish day/month localization for DateCheckFilter.
// Java has no separate DateFilterHelper type — logic lives on pl.DateCheckFilter.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetDayOfWeek ports DateCheckFilter.getDayOfWeek(String).
// Java: dayStr.toLowerCase() then sequential startsWith/equals; throws RuntimeException.
//
//	pon* → MONDAY
//	wt* → TUESDAY
//	śr* → WEDNESDAY  (no ASCII "sr" invent)
//	czw* → THURSDAY
//	pt | piątk* | piątek → FRIDAY  (no ASCII "piatk"/"piatek" invent)
//	sob* → SATURDAY
//	niedz* → SUNDAY
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	day := strings.ToLower(dayStr)
	switch {
	case strings.HasPrefix(day, "pon"):
		return time.Monday, nil
	case strings.HasPrefix(day, "wt"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "śr"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "czw"):
		return time.Thursday, nil
	case day == "pt" || strings.HasPrefix(day, "piątk") || day == "piątek":
		return time.Friday, nil
	case strings.HasPrefix(day, "sob"):
		return time.Saturday, nil
	case strings.HasPrefix(day, "niedz"):
		return time.Sunday, nil
	default:
		// Java: throw new RuntimeException("Could not find day of week for '" + dayStr + "'");
		return 0, fmt.Errorf("Could not find day of week for '%s'", dayStr)
	}
}

// GetDayOfWeekName ports DateCheckFilter.getDayOfWeek(Calendar) —
// date.getDisplayName(DAY_OF_WEEK, LONG, Locale.forLanguageTag("pl")).
func (h *DateFilterHelper) GetDayOfWeekName(t time.Time) string {
	// Java Locale "pl" LONG weekday names (nominative).
	names := []string{"niedziela", "poniedziałek", "wtorek", "środa", "czwartek", "piątek", "sobota"}
	return names[int(t.Weekday())]
}

// GetMonth ports DateCheckFilter.getMonth(String) → time.Month (January=1).
// Java: mon = monthStr.toLowerCase(); genitive mon.equals OR case-sensitive monthStr.equals Roman.
// No nominative (styczeń/…) or ASCII-stripped forms — Java does not accept them.
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := strings.ToLower(monthStr)
	// Java sequential ifs: mon.equals("stycznia") || monthStr.equals("I") etc.
	switch {
	case mon == "stycznia" || monthStr == "I":
		return time.January, nil
	case mon == "lutego" || monthStr == "II":
		return time.February, nil
	case mon == "marca" || monthStr == "III":
		return time.March, nil
	case mon == "kwietnia" || monthStr == "IV":
		return time.April, nil
	case mon == "maja" || monthStr == "V":
		return time.May, nil
	case mon == "czerwca" || monthStr == "VI":
		return time.June, nil
	case mon == "lipca" || monthStr == "VII":
		return time.July, nil
	case mon == "sierpnia" || monthStr == "VIII":
		return time.August, nil
	case mon == "września" || monthStr == "IX":
		return time.September, nil
	case mon == "października" || monthStr == "X":
		return time.October, nil
	case mon == "listopada" || monthStr == "XI":
		return time.November, nil
	case mon == "grudnia" || monthStr == "XII":
		return time.December, nil
	default:
		// Java: throw new RuntimeException("Could not find month '" + monthStr + "'");
		return 0, fmt.Errorf("Could not find month '%s'", monthStr)
	}
}
