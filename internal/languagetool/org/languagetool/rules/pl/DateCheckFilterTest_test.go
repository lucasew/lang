package pl

// Twin of org.languagetool.rules.pl.DateCheckFilterTest
import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Port of DateCheckFilterTest.testGetDayOfWeek
// Java Calendar: Sunday=1, Monday=2, … Saturday=7
func TestDateCheckFilter_GetDayOfWeek(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		{"niedz", 1},
		{"pon", 2},
		{"Pon", 2},
		{"pon.", 2},
		{"poniedziałek", 2},
		{"Poniedziałek", 2},
		{"wtorek", 3},
		{"pt", 6},
		{"piątek", 6},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := f.GetDayOfWeekJava(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// Port of DateCheckFilterTest.testMonth
func TestDateCheckFilter_Month(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		{"I", 1},
		{"XII", 12},
		{"grudnia", 12},
		{"Grudnia", 12},
		{"GRUDNIA", 12},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := f.GetMonth(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// Expanded matrix beyond Java unit test — same getDayOfWeek rules.
func TestDateCheckFilter_GetDayOfWeek_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		{"niedz", 1},
		{"niedziela", 1},
		{"pon", 2},
		{"poniedziałek", 2},
		{"wt", 3},
		{"wtorek", 3},
		{"śr", 4},
		{"środa", 4},
		{"czw", 5},
		{"czwartek", 5},
		{"pt", 6},
		{"piątek", 6},
		{"piątk", 6}, // startsWith("piątk")
		{"sob", 7},
		{"sobota", 7},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := f.GetDayOfWeekJava(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestDateCheckFilter_GetDayOfWeek_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	// Invent/ASCII forms Java does not accept (startsWith is not whole-word).
	for _, in := range []string{"xyz", "sr", "piatek", "piatk", "monday", "friday"} {
		_, err := f.GetDayOfWeekJava(in)
		require.Error(t, err, "expected error for %q", in)
	}
}

func TestDateCheckFilter_GetDayOfWeek_CalendarFormat(t *testing.T) {
	f := NewDateCheckFilter()
	// 2014-08-24..30: Sunday..Saturday
	// 24 Aug 2014 was Sunday
	require.Equal(t, "niedziela", f.GetDayOfWeekName(2014, 8, 24))
	require.Equal(t, "poniedziałek", f.GetDayOfWeekName(2014, 8, 25))
	require.Equal(t, "wtorek", f.GetDayOfWeekName(2014, 8, 26))
	require.Equal(t, "środa", f.GetDayOfWeekName(2014, 8, 27))
	require.Equal(t, "czwartek", f.GetDayOfWeekName(2014, 8, 28))
	require.Equal(t, "piątek", f.GetDayOfWeekName(2014, 8, 29))
	require.Equal(t, "sobota", f.GetDayOfWeekName(2014, 8, 30))
}

// Expanded getMonth matrix: genitive + Roman (case-sensitive Roman).
func TestDateCheckFilter_GetMonth_Matrix(t *testing.T) {
	f := NewDateCheckFilter()
	cases := []struct {
		in   string
		want int
	}{
		{"stycznia", 1}, {"I", 1},
		{"lutego", 2}, {"II", 2},
		{"marca", 3}, {"III", 3},
		{"kwietnia", 4}, {"IV", 4},
		{"maja", 5}, {"V", 5},
		{"czerwca", 6}, {"VI", 6},
		{"lipca", 7}, {"VII", 7},
		{"sierpnia", 8}, {"VIII", 8},
		{"września", 9}, {"IX", 9},
		{"października", 10}, {"X", 10},
		{"listopada", 11}, {"XI", 11},
		{"grudnia", 12}, {"XII", 12},
		// case-insensitive genitive
		{"STYCZNIA", 1},
		{"Sierpnia", 8},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := f.GetMonth(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestDateCheckFilter_GetMonth_Unknown(t *testing.T) {
	f := NewDateCheckFilter()
	// Nominative / ASCII invent / lowercase Roman — Java throws
	for _, in := range []string{
		"styczeń", "styczen", "luty", "marzec", "kwiecień", "maj",
		"czerwiec", "lipiec", "sierpień", "wrzesień", "październik",
		"listopad", "grudzień", "wrzesnia", "pazdziernika",
		"i", "xii", "x", "January", "xyz", "",
	} {
		_, err := f.GetMonth(in)
		require.Error(t, err, "expected error for %q", in)
	}
}

func TestDateCheckFilter_GetMonth_RomanCaseSensitive(t *testing.T) {
	f := NewDateCheckFilter()
	// Java: monthStr.equals("I") — case-sensitive; "i" / "Ii" must fail
	_, err := f.GetMonth("i")
	require.Error(t, err)
	_, err = f.GetMonth("Ii")
	require.Error(t, err)
	_, err = f.GetMonth("xii")
	require.Error(t, err)
	m, err := f.GetMonth("I")
	require.NoError(t, err)
	require.Equal(t, 1, m)
}

// Sanity: FormatDayOfWeek via helper uses Go Weekday index.
func TestDateFilterHelper_GetDayOfWeekName_Index(t *testing.T) {
	h := NewDateFilterHelper()
	// time.Sunday=0 … time.Saturday=6
	require.Equal(t, "niedziela", h.GetDayOfWeekName(time.Date(2014, 8, 24, 0, 0, 0, 0, time.UTC)))
	require.Equal(t, "środa", h.GetDayOfWeekName(time.Date(2014, 8, 27, 0, 0, 0, 0, time.UTC)))
}
