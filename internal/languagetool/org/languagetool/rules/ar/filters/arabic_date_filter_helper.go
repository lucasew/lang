package filters

import (
	"fmt"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ArabicDateFilterHelper ports org.languagetool.rules.ar.filters.ArabicDateFilterHelper
// (package-private in Java; same package in Go).
type ArabicDateFilterHelper struct{}

func NewArabicDateFilterHelper() *ArabicDateFilterHelper { return &ArabicDateFilterHelper{} }

// GetCalendar ports ArabicDateFilterHelper.getCalendar → Calendar.getInstance(Locale.UK).
// Locale selects calendar symbols/first-day; wall clock is local (time.Now).
func (h *ArabicDateFilterHelper) GetCalendar() time.Time {
	return time.Now()
}

// GetDayOfWeek ports ArabicDateFilterHelper.getDayOfWeek(String).
// Java switches on dayStr as-is (no trimSpecialCharacters, no toLowerCase).
func (h *ArabicDateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	switch dayStr {
	case "السبت":
		return time.Saturday, nil
	case "الأحد":
		return time.Sunday, nil
	case "الإثنين":
		return time.Monday, nil
	case "الاثنين":
		return time.Monday, nil
	case "الثلاثاء":
		return time.Tuesday, nil
	case "الأربعاء":
		return time.Wednesday, nil
	case "الخميس":
		return time.Thursday, nil
	case "الجمعة":
		return time.Friday, nil
	default:
		// Java: throw new RuntimeException("No day name found for " + dayStr + "'");
		// (note missing opening quote — upstream quirk)
		return 0, fmt.Errorf("No day name found for %s'", dayStr)
	}
}

// GetMonth ports ArabicDateFilterHelper.getMonth(String) → 1..12 (as time.Month).
// Java: String mon = StringTools.trimSpecialCharacters(monthStr); then switch.
func (h *ArabicDateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	mon := tools.TrimSpecialCharacters(monthStr)
	switch mon {
	// الأشهر العربية بالسريانية
	case "كانون الثاني":
		return time.January, nil
	case "كانون ثاني":
		return time.January, nil
	case "شباط":
		return time.February, nil
	case "آذار":
		return time.March, nil
	case "نيسان":
		return time.April, nil
	case "أيار":
		return time.May, nil
	case "حزيران":
		return time.June, nil
	case "تموز":
		return time.July, nil
	case "آب":
		return time.August, nil
	case "أيلول":
		return time.September, nil
	case "تشرين الأول":
		return time.October, nil
	case "تشرين الثاني":
		return time.November, nil
	case "كانون الأول":
		return time.December, nil
	case "تشرين ثاني":
		return time.November, nil
	case "كانون أول":
		return time.December, nil
	// الأشهر المعربة عن الإنجليزية
	case "يناير":
		return time.January, nil
	case "فبراير":
		return time.February, nil
	case "مارس":
		return time.March, nil
	case "أبريل":
		return time.April, nil
	case "مايو":
		return time.May, nil
	case "يونيو":
		return time.June, nil
	case "يوليو":
		return time.July, nil
	case "أغسطس":
		return time.August, nil
	case "سبتمبر":
		return time.September, nil
	case "أكتوبر":
		return time.October, nil
	case "نوفمبر":
		return time.November, nil
	case "ديسمبر":
		return time.December, nil
	// الأشهر المعربة عن الفرنسية
	case "جانفي":
		return time.January, nil
	case "جانفييه":
		return time.January, nil
	case "فيفري":
		return time.February, nil
	case "أفريل":
		return time.April, nil
	case "ماي":
		return time.May, nil
	case "جوان":
		return time.June, nil
	case "جويلية":
		return time.July, nil
	case "أوت":
		return time.August, nil
	default:
		// Java: throw new RuntimeException("No month name for '" + monthStr + "'");
		return 0, fmt.Errorf("No month name for '%s'", monthStr)
	}
}

// GetDayOfWeekName ports ArabicDateFilterHelper.getDayOfWeekName(int day)
// with Calendar.SUNDAY=1…SATURDAY=7 mapped via Go time.Weekday (Sunday=0…Saturday=6).
func (h *ArabicDateFilterHelper) GetDayOfWeekName(day time.Weekday) string {
	switch day {
	case time.Saturday:
		return "السبت"
	case time.Sunday:
		return "الأحد"
	case time.Monday:
		return "الإثنين"
	case time.Tuesday:
		return "الثلاثاء"
	case time.Wednesday:
		return "الأربعاء"
	case time.Thursday:
		return "الخميس"
	case time.Friday:
		return "الجمعة"
	default:
		return "غير محدد"
	}
}
