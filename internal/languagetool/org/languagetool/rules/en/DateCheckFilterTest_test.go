package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/DateCheckFilterTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of DateCheckFilterTest.testGetDayOfWeek — every Java assertThat.
func TestDateCheckFilter_GetDayOfWeek(t *testing.T) {
	f := NewDateCheckFilter()
	// Java Calendar: Sunday=1 … Saturday=7
	cases := []struct {
		in   string
		want int
	}{
		{"Sun", 1},
		{"Mon", 2},
		{"mon", 2},
		{"Mon.", 2},
		{"Monday", 2},
		{"monday", 2},
		{"Tue", 3},
		{"Fri", 6},
		{"Fr", 6},
		{"Saturday", 7},
	}
	for _, tc := range cases {
		d, err := f.GetDayOfWeekJava(tc.in)
		require.NoError(t, err, "day %q", tc.in)
		require.Equal(t, tc.want, d, "day %q", tc.in)
	}
}

// Port of DateCheckFilterTest.testMonth — every Java assertThat.
func TestDateCheckFilter_Month(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		{"jan", 1},
		{"dec", 12},
		{"december", 12},
		{"December", 12},
		{"DECEMBER", 12},
	}
	for _, tc := range cases {
		m, err := f.GetMonth(tc.in)
		require.NoError(t, err, "month %q", tc.in)
		require.Equal(t, tc.want, m, "month %q", tc.in)
	}
}
