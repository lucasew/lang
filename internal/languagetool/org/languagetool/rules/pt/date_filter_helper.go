package pt

import (
	"fmt"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DateFilterHelper ports org.languagetool.rules.pt.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

// GetCalendar ports DateFilterHelper.getCalendar → Calendar.getInstance(Locale.UK).
// Locale.UK selects calendar symbols/first-day; wall clock uses the JVM/OS default TZ
// (Go: time.Now in Local).
func (h *DateFilterHelper) GetCalendar() time.Time {
	return time.Now()
}

// GetDayOfWeek ports DateFilterHelper.getDayOfWeek(String).
// Prefix match after trimSpecialCharacters + toLowerCase (Java).
// Go time.Weekday Sunday=0 maps to Java Calendar.SUNDAY=1 for the same wall day
// (callers that need Java ints add 1; see DateCheckFilter.GetDayOfWeekJava).
func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	// Java: StringTools.trimSpecialCharacters(dayStr).toLowerCase()
	day := strings.ToLower(tools.TrimSpecialCharacters(dayStr))
	switch {
	case strings.HasPrefix(day, "dom"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "seg"):
		return time.Monday, nil
	case strings.HasPrefix(day, "ter"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "qua"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "qui"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "sex"):
		return time.Friday, nil
	// Java only: startsWith("sáb") — no unaccented "sab" invent
	case strings.HasPrefix(day, "sáb"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

// GetDayOfWeekName ports DateFilterHelper.getDayOfWeek(Calendar).
// Java gets LONG English names via Locale.UK, then maps to PT long forms.
func (h *DateFilterHelper) GetDayOfWeekName(t time.Time) string {
	// Java: date.getDisplayName(Calendar.DAY_OF_WEEK, Calendar.LONG, Locale.UK)
	englishNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	englishDay := englishNames[int(t.Weekday())]
	switch englishDay {
	case "Sunday":
		return "domingo"
	case "Monday":
		return "segunda-feira"
	case "Tuesday":
		return "terça-feira"
	case "Wednesday":
		return "quarta-feira"
	case "Thursday":
		return "quinta-feira"
	case "Friday":
		return "sexta-feira"
	case "Saturday":
		return "sábado"
	default:
		// Java returns "" if englishDay is not in the switch (normal Calendar never hits this).
		return ""
	}
}

// GetMonth ports DateFilterHelper.getMonth(String) → 1..12 (as time.Month).
func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	// Java: StringTools.trimSpecialCharacters(monthStr).toLowerCase()
	mon := strings.ToLower(tools.TrimSpecialCharacters(monthStr))
	switch {
	case strings.HasPrefix(mon, "jan"):
		return time.January, nil
	case strings.HasPrefix(mon, "fev"):
		return time.February, nil
	case strings.HasPrefix(mon, "mar"):
		return time.March, nil
	case strings.HasPrefix(mon, "abr"):
		return time.April, nil
	case strings.HasPrefix(mon, "mai"):
		return time.May, nil
	case strings.HasPrefix(mon, "jun"):
		return time.June, nil
	case strings.HasPrefix(mon, "jul"):
		return time.July, nil
	case strings.HasPrefix(mon, "ago"):
		return time.August, nil
	case strings.HasPrefix(mon, "set"):
		return time.September, nil
	case strings.HasPrefix(mon, "out"):
		return time.October, nil
	case strings.HasPrefix(mon, "nov"):
		return time.November, nil
	case strings.HasPrefix(mon, "dez"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
