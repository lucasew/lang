package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidWordFilter(t *testing.T) {
	f := NewValidWordFilter()
	// default: always misspelled → keep match
	require.True(t, f.Accept("vielleicht", "der"))
	f.IsMisspelled = func(w string) bool {
		return w != "Promotionsstudierende"
	}
	require.False(t, f.Accept("Promotions", "Studierende")) // known good joined form
	require.True(t, f.Accept("foo", "Bar"))
}
