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
