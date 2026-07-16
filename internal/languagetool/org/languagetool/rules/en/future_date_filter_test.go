package en

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFutureDateFilter(t *testing.T) {
	f := NewFutureDateFilter()
	require.False(t, f.IsFuture(1999, 12, 31))
	require.True(t, f.IsFuture(time.Now().UTC().Year()+2, 1, 1))
}
