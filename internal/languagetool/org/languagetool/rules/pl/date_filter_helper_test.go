package pl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateFilterHelper_GetDayOfWeek(t *testing.T) {
	h := NewDateFilterHelper()
	cases := []struct {
		in   string
		want time.Weekday
	}{
		{"poniedziałek", time.Monday},
		{"Pon", time.Monday},
		{"wtorek", time.Tuesday},
		{"środa", time.Wednesday},
		{"czwartek", time.Thursday},
		{"pt", time.Friday},
		{"piątek", time.Friday},
		{"sobota", time.Saturday},
		{"niedz", time.Sunday},
	}
	for _, tc := range cases {
		got, err := h.GetDayOfWeek(tc.in)
		require.NoError(t, err, tc.in)
		require.Equal(t, tc.want, got, tc.in)
	}
}

func TestDateFilterHelper_GetDayOfWeek_NoASCIIInvent(t *testing.T) {
	h := NewDateFilterHelper()
	for _, in := range []string{"sr", "piatk", "piatek"} {
		_, err := h.GetDayOfWeek(in)
		require.Error(t, err, in)
	}
}

func TestDateFilterHelper_GetMonth(t *testing.T) {
	h := NewDateFilterHelper()
	m, err := h.GetMonth("stycznia")
	require.NoError(t, err)
	require.Equal(t, time.January, m)
	m, err = h.GetMonth("III")
	require.NoError(t, err)
	require.Equal(t, time.March, m)
	m, err = h.GetMonth("grudnia")
	require.NoError(t, err)
	require.Equal(t, time.December, m)
}

func TestDateFilterHelper_GetMonth_NoNominativeInvent(t *testing.T) {
	h := NewDateFilterHelper()
	for _, in := range []string{"styczeń", "luty", "sierpień", "i", "xii"} {
		_, err := h.GetMonth(in)
		require.Error(t, err, in)
	}
}
