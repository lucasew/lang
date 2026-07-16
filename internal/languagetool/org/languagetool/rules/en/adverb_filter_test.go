package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdverbFilter(t *testing.T) {
	f := NewAdverbFilter()
	require.Equal(t, "good performance", f.Suggest("well", "performance"))
	require.Equal(t, "simple solution", f.Suggest("simply", "solution"))
	require.Equal(t, "easy task", f.Suggest("easily", "task"))
	require.Equal(t, "", f.Suggest("fast", "car")) // same form
	require.Equal(t, "", f.Suggest("notanadverb", "x"))
	require.Equal(t, "second attempt", f.Suggest("twice", "attempt"))
}
