package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDashPrefixes_JavaFilters(t *testing.T) {
	m := loadDashPrefixes()
	require.NotEmpty(t, m)
	_, hasBlock := m["блок"]
	_, hasRating := m["рейтинг"]
	require.False(t, hasBlock)
	require.False(t, hasRating)
	// known good keys
	require.Contains(t, m, "штаб")
	require.Equal(t, "", m["штаб"])
	require.Contains(t, m, "міні")
	require.Equal(t, ":alt", m["міні"])
	require.Contains(t, m, "екс")
	require.Equal(t, ":alt", m["екс"])
	// uppercase / latin filtered by ALL_LOWER
	_, hasAI := m["AI"]
	require.False(t, hasAI)
	// :bad stripped
	for _, v := range m {
		require.NotContains(t, v, ":bad")
	}
	// тайм is NOT in map (special-cased in match)
	_, hasTaim := m["тайм"]
	require.False(t, hasTaim)
}
