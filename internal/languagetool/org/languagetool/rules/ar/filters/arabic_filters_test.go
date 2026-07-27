package filters

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestArabicDateFilterHelper(t *testing.T) {
	h := NewArabicDateFilterHelper()
	wd, err := h.GetDayOfWeek("الجمعة")
	require.NoError(t, err)
	require.Equal(t, time.Friday, wd)
	_, err = h.GetDayOfWeek("notaday")
	require.Error(t, err)
	require.Equal(t, "No day name found for notaday'", err.Error())

	// both Monday spellings
	wd, err = h.GetDayOfWeek("الإثنين")
	require.NoError(t, err)
	require.Equal(t, time.Monday, wd)
	wd, err = h.GetDayOfWeek("الاثنين")
	require.NoError(t, err)
	require.Equal(t, time.Monday, wd)

	m, err := h.GetMonth("يناير")
	require.NoError(t, err)
	require.Equal(t, time.January, m)
	m, err = h.GetMonth("شباط")
	require.NoError(t, err)
	require.Equal(t, time.February, m)
	m, err = h.GetMonth("جويلية")
	require.NoError(t, err)
	require.Equal(t, time.July, m)
	_, err = h.GetMonth("notamonth")
	require.Error(t, err)
	require.Equal(t, "No month name for 'notamonth'", err.Error())

	// helper name path (int overload): Monday with hamza
	require.Equal(t, "الإثنين", h.GetDayOfWeekName(time.Monday))
}

func TestArabicDateCheckFilter(t *testing.T) {
	f := NewArabicDateCheckFilter()
	j, err := f.GetDayOfWeekJava("الأحد")
	require.NoError(t, err)
	require.Equal(t, 1, j) // Sunday=1 in Java
	mi, err := f.GetMonth("ديسمبر")
	require.NoError(t, err)
	require.Equal(t, 12, mi)
	// getDayOfWeek(Calendar) → ar Locale LONG: Monday = "الاثنين" (no hamza)
	require.Equal(t, "الاثنين", f.GetDayOfWeekName(2024, 1, 1))
	// getDayOfWeek(int) → helper: Monday = "الإثنين" (with hamza)
	require.Equal(t, "الإثنين", f.GetDayOfWeekNameFromJavaDay(2))
	_ = NewArabicDMYDateCheckFilter() // DMY not claimed ready in this sector
}

func TestArabicNumberPhraseFilter(t *testing.T) {
	got := SuggestionsForNumericPhrase("3", false)
	require.Equal(t, []string{"ثلاثة"}, got)
	got = PrepareSuggestion("5", "في", false)
	require.Equal(t, []string{"في خمسة"}, got)
	got = PrepareSuggestionWithUnit("2", "", "دينار", "raf3", false)
	require.Len(t, got, 1)
	require.Contains(t, got[0], "اثنان")
	require.Contains(t, got[0], "ديناران")
	require.Equal(t, "jar", InflectionFromPrevious("بشيء"))
	require.Equal(t, "", InflectionFromPrevious("شيء"))
}
