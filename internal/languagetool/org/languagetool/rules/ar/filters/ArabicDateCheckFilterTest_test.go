package filters

// Twin of ArabicDateCheckFilterTest (Java king).
import (
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestArabicDateCheckFilter_Accept(t *testing.T) {
	f := NewArabicDateCheckFilter()
	match := rules.NewRuleMatch(nil, nil, 0, 10, "message")
	// 2022-03-12 was Saturday (السبت) — correct → nil
	require.Nil(t, f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "السبت",
	}, -1, nil, nil))
	// claim الأحد (Sunday) — incorrect → non-nil
	require.NotNil(t, f.AcceptRuleMatch(match, map[string]string{
		"year": "2022", "month": "3", "day": "12", "weekDay": "الأحد",
	}, -1, nil, nil))
}

func TestArabicDateCheckFilter_AcceptIncompleteArgs(t *testing.T) {
	f := NewArabicDateCheckFilter()
	match := rules.NewRuleMatch(nil, nil, 0, 10, "message")
	// Java: IllegalArgumentException when weekDay missing
	require.Panics(t, func() {
		f.AcceptRuleMatch(match, map[string]string{
			"year": "2022", "month": "3", "day": "12",
		}, -1, nil, nil)
	})
}

func TestArabicDateCheckFilter_GetDayOfWeek1(t *testing.T) {
	f := NewArabicDateCheckFilter()
	// Java Calendar: SUNDAY=1 … SATURDAY=7 via getDayOfWeek(String)
	cases := []struct {
		name string
		java int
	}{
		{"الأحد", 1},
		{"الإثنين", 2},
		{"الثلاثاء", 3},
		{"الأربعاء", 4},
		{"الخميس", 5},
		{"الجمعة", 6},
		{"السبت", 7},
	}
	for _, c := range cases {
		j, err := f.GetDayOfWeekJava(c.name)
		require.NoError(t, err, c.name)
		require.Equal(t, c.java, j, c.name)
	}
	// Java also accepts الاثنين (without hamza) → MONDAY
	j, err := f.GetDayOfWeekJava("الاثنين")
	require.NoError(t, err)
	require.Equal(t, 2, j)

	// inverse: Java filter.getDayOfWeek(Calendar.X) is the int overload → helper names
	// (Monday = "الإثنين" with hamza — NOT ar Locale displayName)
	require.Equal(t, "الأحد", f.GetDayOfWeekNameFromJavaDay(1))   // Calendar.SUNDAY
	require.Equal(t, "الإثنين", f.GetDayOfWeekNameFromJavaDay(2)) // Calendar.MONDAY
	require.Equal(t, "الثلاثاء", f.GetDayOfWeekNameFromJavaDay(3))
	require.Equal(t, "الأربعاء", f.GetDayOfWeekNameFromJavaDay(4))
	require.Equal(t, "الخميس", f.GetDayOfWeekNameFromJavaDay(5))
	require.Equal(t, "الجمعة", f.GetDayOfWeekNameFromJavaDay(6))
	require.Equal(t, "السبت", f.GetDayOfWeekNameFromJavaDay(7))

	// helper twin of getDayOfWeekName
	h := NewArabicDateFilterHelper()
	require.Equal(t, "الأحد", h.GetDayOfWeekName(time.Sunday))
	require.Equal(t, "الإثنين", h.GetDayOfWeekName(time.Monday))
	require.Equal(t, "الثلاثاء", h.GetDayOfWeekName(time.Tuesday))
	require.Equal(t, "الأربعاء", h.GetDayOfWeekName(time.Wednesday))
	require.Equal(t, "الخميس", h.GetDayOfWeekName(time.Thursday))
	require.Equal(t, "الجمعة", h.GetDayOfWeekName(time.Friday))
	require.Equal(t, "السبت", h.GetDayOfWeekName(time.Saturday))
}

func TestArabicDateCheckFilter_GetDayOfWeek2(t *testing.T) {
	f := NewArabicDateCheckFilter()
	// Java: Calendar.set(2022, MARCH, 25/26); getDayOfWeek(calendar)
	// → date.getDisplayName(DAY_OF_WEEK, LONG, Locale.forLanguageTag("ar"))
	// 2022-03-25 Friday, 2022-03-26 Saturday
	require.Equal(t, "الجمعة", f.GetDayOfWeekName(2022, 3, 25))
	require.Equal(t, "السبت", f.GetDayOfWeekName(2022, 3, 26))
	// Monday via ar Locale LONG is "الاثنين" (no hamza) — not helper "الإثنين"
	require.Equal(t, "الاثنين", f.GetDayOfWeekName(2024, 1, 1))
}

func TestArabicDateCheckFilter_GetMonth(t *testing.T) {
	f := NewArabicDateCheckFilter()
	cases := map[string]int{
		"جانفي":       1,
		"جانفييه":     1,
		"يناير":       1,
		"ديسمبر":      12,
		"كانون الأول": 12,
		"كانون أول":   12,
		"أبريل":       4,
		"نيسان":       4,
	}
	for name, want := range cases {
		got, err := f.GetMonth(name)
		require.NoError(t, err, name)
		require.Equal(t, want, got, name)
	}
}
