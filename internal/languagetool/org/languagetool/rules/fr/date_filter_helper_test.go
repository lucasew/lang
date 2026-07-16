package fr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	d, err := h.GetDayOfWeek("lundi")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	m, err := h.GetMonth("février")
	require.NoError(t, err)
	require.Equal(t, time.February, m)
	m, err = h.GetMonth("juin")
	require.NoError(t, err)
	require.Equal(t, time.June, m)
}
