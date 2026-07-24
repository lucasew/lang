package en

import (
	"fmt"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DateFilterHelper ports org.languagetool.rules.en.DateFilterHelper.
type DateFilterHelper struct{}

func NewDateFilterHelper() *DateFilterHelper { return &DateFilterHelper{} }

func (h *DateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	// Java: StringTools.trimSpecialCharacters(dayStr).toLowerCase()
	day := strings.ToLower(tools.TrimSpecialCharacters(dayStr))
	switch {
	case strings.HasPrefix(day, "su"):
		return time.Sunday, nil
	case strings.HasPrefix(day, "mo"):
		return time.Monday, nil
	case strings.HasPrefix(day, "tu"):
		return time.Tuesday, nil
	case strings.HasPrefix(day, "we"):
		return time.Wednesday, nil
	case strings.HasPrefix(day, "th"):
		return time.Thursday, nil
	case strings.HasPrefix(day, "fr"):
		return time.Friday, nil
	case strings.HasPrefix(day, "sa"):
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("could not find day of week for %q", dayStr)
	}
}

func (h *DateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	// Java: StringTools.trimSpecialCharacters(monthStr).toLowerCase()
	mon := strings.ToLower(tools.TrimSpecialCharacters(monthStr))
	switch {
	case strings.HasPrefix(mon, "jan"):
		return time.January, nil
	case strings.HasPrefix(mon, "feb"):
		return time.February, nil
	case strings.HasPrefix(mon, "mar"):
		return time.March, nil
	case strings.HasPrefix(mon, "apr"):
		return time.April, nil
	case strings.HasPrefix(mon, "may"):
		return time.May, nil
	case strings.HasPrefix(mon, "jun"):
		return time.June, nil
	case strings.HasPrefix(mon, "jul"):
		return time.July, nil
	case strings.HasPrefix(mon, "aug"):
		return time.August, nil
	case strings.HasPrefix(mon, "sep"):
		return time.September, nil
	case strings.HasPrefix(mon, "oct"):
		return time.October, nil
	case strings.HasPrefix(mon, "nov"):
		return time.November, nil
	case strings.HasPrefix(mon, "dec"):
		return time.December, nil
	default:
		return 0, fmt.Errorf("could not find month %q", monthStr)
	}
}
