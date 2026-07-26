package uk

// Twin of org.languagetool.rules.uk.DateCheckFilterTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter_GetDayOfWeek(t *testing.T) {
	filter := NewDateCheckFilter()
	// Java Calendar: Sunday=1 … Saturday=7
	assertDay := func(s string, want int) {
		t.Helper()
		got, err := filter.GetDayOfWeekJava(s)
		require.NoError(t, err, s)
		require.Equal(t, want, got, s)
	}
	// Java DateCheckFilterTest.testGetDayOfWeek
	assertDay("Нед", 1)
	assertDay("Пон", 2)
	assertDay("пон", 2)
	assertDay("Понед.", 2)
	assertDay("Понеділок", 2)
	assertDay("понеділок", 2)
	assertDay("Вт", 3)
	assertDay("Сер", 4)
	assertDay("П'ят", 6)
	assertDay("Суб", 7)
}

func TestDateCheckFilter_Month(t *testing.T) {
	filter := NewDateCheckFilter()
	assertMonth := func(s string, want int) {
		t.Helper()
		got, err := filter.GetMonth(s)
		require.NoError(t, err, s)
		require.Equal(t, want, got, s)
	}
	// Java DateCheckFilterTest.testMonth
	assertMonth("січ", 1)
	assertMonth("гру", 12)
	assertMonth("грудень", 12)
	assertMonth("Грудень", 12)
	assertMonth("ГРУДЕНЬ", 12)
}

// Java-quirk coverage: second startsWith("ли") for November is dead code —
// "листопад" (November) matches July first.
func TestDateCheckFilter_Month_JulyPrefixWins(t *testing.T) {
	filter := NewDateCheckFilter()
	m, err := filter.GetMonth("ли")
	require.NoError(t, err)
	require.Equal(t, 7, m)
	m, err = filter.GetMonth("липень")
	require.NoError(t, err)
	require.Equal(t, 7, m)
	m, err = filter.GetMonth("листопад")
	require.NoError(t, err)
	require.Equal(t, 7, m, "Java dead-code quirk: ли* is always July")
}
