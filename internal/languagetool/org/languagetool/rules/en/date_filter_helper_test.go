package en

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	d, err := h.GetDayOfWeek("Monday")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	m, err := h.GetMonth("September")
	require.NoError(t, err)
	require.Equal(t, time.September, m)
}

func TestDateFilterHelper_TrimSpecialCharacters(t *testing.T) {
	h := NewDateFilterHelper()
	// Java StringTools.trimSpecialCharacters strips soft hyphen, keeps letters
	m, err := h.GetMonth("Sep\u00ADtember")
	require.NoError(t, err)
	require.Equal(t, time.September, m)
	d, err := h.GetDayOfWeek("Mon\u00ADday")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
}
