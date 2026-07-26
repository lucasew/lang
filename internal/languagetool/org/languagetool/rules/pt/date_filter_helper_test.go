package pt

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
		// short prefixes (Java startsWith)
		{"dom", time.Sunday},
		{"seg", time.Monday},
		{"ter", time.Tuesday},
		{"qua", time.Wednesday},
		{"qui", time.Thursday},
		{"sex", time.Friday},
		{"sáb", time.Saturday},
		// longer / full forms
		{"domingo", time.Sunday},
		{"segunda", time.Monday},
		{"segunda-feira", time.Monday},
		{"terça", time.Tuesday},
		{"terça-feira", time.Tuesday},
		{"quarta", time.Wednesday},
		{"quarta-feira", time.Wednesday},
		{"quinta", time.Thursday},
		{"quinta-feira", time.Thursday},
		{"sexta", time.Friday},
		{"sexta-feira", time.Friday},
		{"sábado", time.Saturday},
		// mixed case after toLowerCase
		{"Domingo", time.Sunday},
		{"SEGUNDA", time.Monday},
		{"Sábado", time.Saturday},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := h.GetDayOfWeek(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestDateFilterHelper_GetDayOfWeek_UnaccentedSaturdayErrors(t *testing.T) {
	h := NewDateFilterHelper()
	// Java only startsWith("sáb") — unaccented "sab" is invent, must fail
	for _, in := range []string{"sab", "sabado", "SAB", "Sabado"} {
		_, err := h.GetDayOfWeek(in)
		require.Error(t, err, "expected error for %q", in)
	}
}

func TestDateFilterHelper_GetDayOfWeek_Unknown(t *testing.T) {
	h := NewDateFilterHelper()
	_, err := h.GetDayOfWeek("xyz")
	require.Error(t, err)
	_, err = h.GetDayOfWeek("")
	require.Error(t, err)
}

func TestDateFilterHelper_GetDayOfWeek_TrimSpecialCharacters(t *testing.T) {
	h := NewDateFilterHelper()
	// soft hyphen U+00AD stripped by StringTools.trimSpecialCharacters
	d, err := h.GetDayOfWeek("seg\u00ADunda")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	d, err = h.GetDayOfWeek("sáb\u00ADado")
	require.NoError(t, err)
	require.Equal(t, time.Saturday, d)
}

func TestDateFilterHelper_GetMonth(t *testing.T) {
	h := NewDateFilterHelper()
	cases := []struct {
		in   string
		want time.Month
	}{
		{"jan", time.January},
		{"fev", time.February},
		{"mar", time.March},
		{"abr", time.April},
		{"mai", time.May},
		{"jun", time.June},
		{"jul", time.July},
		{"ago", time.August},
		{"set", time.September},
		{"out", time.October},
		{"nov", time.November},
		{"dez", time.December},
		// longer forms
		{"janeiro", time.January},
		{"fevereiro", time.February},
		{"março", time.March},
		{"abril", time.April},
		{"maio", time.May},
		{"junho", time.June},
		{"julho", time.July},
		{"agosto", time.August},
		{"setembro", time.September},
		{"outubro", time.October},
		{"novembro", time.November},
		{"dezembro", time.December},
		{"FEV", time.February},
		{"Dezembro", time.December},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := h.GetMonth(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestDateFilterHelper_GetMonth_Unknown(t *testing.T) {
	h := NewDateFilterHelper()
	_, err := h.GetMonth("xyz")
	require.Error(t, err)
	// English month invent — not in Java PT helper
	_, err = h.GetMonth("february")
	require.Error(t, err)
	_, err = h.GetMonth("")
	require.Error(t, err)
}

func TestDateFilterHelper_GetMonth_TrimSpecialCharacters(t *testing.T) {
	h := NewDateFilterHelper()
	m, err := h.GetMonth("fev\u00ADereiro")
	require.NoError(t, err)
	require.Equal(t, time.February, m)
	m, err = h.GetMonth("set\u00ADembro")
	require.NoError(t, err)
	require.Equal(t, time.September, m)
}

func TestDateFilterHelper_GetDayOfWeekName(t *testing.T) {
	h := NewDateFilterHelper()
	// Fixed UTC dates with known weekdays (same anchors as DateCheckFilter tests).
	cases := []struct {
		t    time.Time
		want string
	}{
		{time.Date(2014, 8, 24, 0, 0, 0, 0, time.UTC), "domingo"},      // Sunday
		{time.Date(2014, 8, 25, 0, 0, 0, 0, time.UTC), "segunda-feira"}, // Monday
		{time.Date(2014, 8, 26, 0, 0, 0, 0, time.UTC), "terça-feira"},   // Tuesday
		{time.Date(2014, 8, 27, 0, 0, 0, 0, time.UTC), "quarta-feira"},  // Wednesday
		{time.Date(2014, 8, 28, 0, 0, 0, 0, time.UTC), "quinta-feira"},  // Thursday
		{time.Date(2014, 8, 29, 0, 0, 0, 0, time.UTC), "sexta-feira"},   // Friday
		{time.Date(2014, 8, 23, 0, 0, 0, 0, time.UTC), "sábado"},        // Saturday
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			require.Equal(t, tc.want, h.GetDayOfWeekName(tc.t))
			require.Equal(t, tc.t.Weekday().String() != "", true)
		})
	}
}

func TestDateFilterHelper_GetCalendar(t *testing.T) {
	h := NewDateFilterHelper()
	before := time.Now().Add(-time.Second)
	got := h.GetCalendar()
	after := time.Now().Add(time.Second)
	require.False(t, got.Before(before))
	require.False(t, got.After(after))
}

// Smoke kept for continuity with earlier partial twin tests.
func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	d, err := h.GetDayOfWeek("segunda")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	m, err := h.GetMonth("fevereiro")
	require.NoError(t, err)
	require.Equal(t, time.February, m)
}

func TestDateFilterHelper_JavaSaturdayPrefix(t *testing.T) {
	h := NewDateFilterHelper()
	_, err := h.GetDayOfWeek("sabado")
	require.Error(t, err)
	d, err := h.GetDayOfWeek("sábado")
	require.NoError(t, err)
	require.Equal(t, time.Saturday, d)
}

func TestDateFilterHelper_TrimSpecialCharacters(t *testing.T) {
	h := NewDateFilterHelper()
	m, err := h.GetMonth("fev\u00ADereiro")
	require.NoError(t, err)
	require.Equal(t, time.February, m)
}
