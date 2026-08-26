package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestionFilter(t *testing.T) {
	f := NewSuggestionFilter(func(filled string) bool {
		return strings.Contains(filled, "bad")
	})
	got := f.Filter([]string{"good", "bad", "ok"}, "This is {}.")
	require.Equal(t, []string{"good", "ok"}, got)
}
