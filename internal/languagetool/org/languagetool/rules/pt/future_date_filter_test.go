package pt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFutureDateFilter_IsFuture(t *testing.T) {
	f := NewFutureDateFilter()
	require.False(t, f.IsFuture(1990, 1, 1))
	require.True(t, f.IsFuture(time.Now().UTC().Year()+3, 3, 1))
}
