package fr

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterrogativeVerbFilter_DesiredPostag(t *testing.T) {
	f := NewInterrogativeVerbFilter()
	require.Contains(t, f.DesiredPostagForPronoun("je"), "1 s")
	require.Contains(t, f.DesiredPostagForPronoun("tu"), "imp")
	require.Contains(t, f.DesiredPostagForPronoun("ils"), "3 p")
	require.Empty(t, f.DesiredPostagForPronoun("xyz"))
}

func TestInterrogativeVerbFilter_Filter(t *testing.T) {
	f := NewInterrogativeVerbFilter()
	got := f.FilterByDesiredPOS([]string{"mange", "manges", "mangeons"}, "2", func(form, re string) bool {
		return strings.HasSuffix(form, "es")
	})
	require.Equal(t, []string{"manges"}, got)
}
