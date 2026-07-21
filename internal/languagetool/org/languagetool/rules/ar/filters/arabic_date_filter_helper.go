package filters

import (
	"fmt"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ArabicDateFilterHelper ports org.languagetool.rules.ar.filters.ArabicDateFilterHelper.
type ArabicDateFilterHelper struct{}

func NewArabicDateFilterHelper() *ArabicDateFilterHelper { return &ArabicDateFilterHelper{} }

func (h *ArabicDateFilterHelper) GetDayOfWeek(dayStr string) (time.Weekday, error) {
	// Java: switch on dayStr as-is (no trimSpecialCharacters)
	switch dayStr {
	case "السبت":
		return time.Saturday, nil
	case "الأحد":
		return time.Sunday, nil
	case "الإثنين", "الاثنين":
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
		return 0, fmt.Errorf("no day name found for %q", dayStr)
	}
}

func (h *ArabicDateFilterHelper) GetMonth(monthStr string) (time.Month, error) {
	// Java: String mon = StringTools.trimSpecialCharacters(monthStr);
	mon := tools.TrimSpecialCharacters(monthStr)
	switch mon {
	// Syriac-style Arabic months
	case "كانون الثاني", "كانون ثاني", "يناير", "جانفي", "جانفييه":
		return time.January, nil
	case "شباط", "فبراير", "فيفري":
		return time.February, nil
	case "آذار", "مارس":
		return time.March, nil
	case "نيسان", "أبريل", "أفريل":
		return time.April, nil
	case "أيار", "مايو", "ماي":
		return time.May, nil
	case "حزيران", "يونيو", "جوان":
		return time.June, nil
	case "تموز", "يوليو", "جويلية":
		return time.July, nil
	case "آب", "أغسطس", "أوت":
		return time.August, nil
	case "أيلول", "سبتمبر":
		return time.September, nil
	case "تشرين الأول", "أكتوبر":
		return time.October, nil
	case "تشرين الثاني", "تشرين ثاني", "نوفمبر":
		return time.November, nil
	case "كانون الأول", "كانون أول", "ديسمبر":
		return time.December, nil
	default:
		return 0, fmt.Errorf("no month name for %q", monthStr)
	}
}

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
