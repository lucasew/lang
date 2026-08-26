package suggestions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentitySuggestionsOrderer(t *testing.T) {
	o := IdentitySuggestionsOrderer{}
	require.False(t, o.IsMlAvailable())
	got := OrderSuggestionsUsingModel(o, []string{"b", "a"}, "x", nil, 0)
	require.Equal(t, []string{"b", "a"}, got)
}

func TestEditDistanceSuggestionsOrderer(t *testing.T) {
	o := EditDistanceSuggestionsOrderer{}
	got := OrderSuggestionsUsingModel(o, []string{"helo", "hello", "hallo"}, "hello", nil, 0)
	require.Equal(t, "hello", got[0])
}
