package de

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	d, err := h.GetDayOfWeek("Montag")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	d, err = h.GetDayOfWeek("So")
	require.NoError(t, err)
	require.Equal(t, time.Sunday, d)
	m, err := h.GetMonth("Januar")
	require.NoError(t, err)
	require.Equal(t, time.January, m)
	m, err = h.GetMonth("März")
	require.NoError(t, err)
	require.Equal(t, time.March, m)
	_, err = h.GetMonth("xyz")
	require.Error(t, err)
}
