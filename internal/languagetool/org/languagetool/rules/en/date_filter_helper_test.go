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
