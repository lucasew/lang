package it

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Colocated helper twins (same localization as DateCheckFilter methods).
func TestDateFilterHelper_Standalone(t *testing.T) {
	h := NewDateFilterHelper()
	d, err := h.GetDayOfWeek("lunedì")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	// full prefix matrix sample
	for _, tc := range []struct {
		in   string
		want time.Weekday
	}{
		{"do", time.Sunday},
		{"lu", time.Monday},
		{"ma", time.Tuesday},
		{"me", time.Wednesday},
		{"gi", time.Thursday},
		{"ve", time.Friday},
		{"sa", time.Saturday},
	} {
		wd, err := h.GetDayOfWeek(tc.in)
		require.NoError(t, err, "day %q", tc.in)
		require.Equal(t, tc.want, wd, "day %q", tc.in)
	}
	m, err := h.GetMonth("gennaio")
	require.NoError(t, err)
	require.Equal(t, time.January, m)
	for _, tc := range []struct {
		in   string
		want time.Month
	}{
		{"gen", time.January},
		{"feb", time.February},
		{"mar", time.March},
		{"apr", time.April},
		{"mag", time.May},
		{"giu", time.June},
		{"lug", time.July},
		{"ago", time.August},
		{"set", time.September},
		{"ott", time.October},
		{"nov", time.November},
		{"dic", time.December},
	} {
		got, err := h.GetMonth(tc.in)
		require.NoError(t, err, "month %q", tc.in)
		require.Equal(t, tc.want, got, "month %q", tc.in)
	}
}
