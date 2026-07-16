package pl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateFilterHelper(t *testing.T) {
	h := NewDateFilterHelper()
	d, err := h.GetDayOfWeek("poniedziałek")
	require.NoError(t, err)
	require.Equal(t, time.Monday, d)
	m, err := h.GetMonth("stycznia")
	require.NoError(t, err)
	require.Equal(t, time.January, m)
	m, err = h.GetMonth("III")
	require.NoError(t, err)
	require.Equal(t, time.March, m)
}
