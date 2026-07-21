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
	// Java Calendar: SUNDAY=1 … SATURDAY=7
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
	// inverse: Java getDayOfWeek(Calendar.X) via helper names
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
	// Java Calendar.MARCH is month index 2 → calendar date March 25/26 2022
	// 2022-03-25 Friday, 2022-03-26 Saturday
	require.Equal(t, "الجمعة", f.GetDayOfWeekName(2022, 3, 25))
	require.Equal(t, "السبت", f.GetDayOfWeekName(2022, 3, 26))
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
