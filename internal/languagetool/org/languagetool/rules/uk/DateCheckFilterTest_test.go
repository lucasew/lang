package uk

// Twin of DateCheckFilterTest
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
	assertMonth("січ", 1)
	assertMonth("гру", 12)
	assertMonth("грудень", 12)
	assertMonth("Грудень", 12)
	assertMonth("ГРУДЕНЬ", 12)
}
