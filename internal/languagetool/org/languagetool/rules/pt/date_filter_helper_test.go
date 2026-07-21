package pt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
	// Java only startsWith("sáb") — unaccented "sab" is invent, must fail
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
