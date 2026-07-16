package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpanishNumberInWordFilter(t *testing.T) {
	f := NewSpanishNumberInWordFilter()
	require.Nil(t, f.Suggestions("hola"))
	require.Equal(t, []string{"cas"}, f.Suggestions("cas4"))
	require.Equal(t, []string{"todo", "td"}, f.Suggestions("t0d0"))
	// all-digit word: only 0→o form (digit-stripped is empty)
	require.Equal(t, []string{"o23"}, f.Suggestions("023"))
}
