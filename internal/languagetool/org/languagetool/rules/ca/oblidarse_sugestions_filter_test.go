package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOblidarseSugestionsFilter(t *testing.T) {
	f := NewOblidarseSugestionsFilter()
	require.True(t, f.NeedsApostrophe("oblidat"))
	require.False(t, f.NeedsApostrophe("passat"))
	require.Equal(t, "m'", f.ReflexivePrefix("1S", true, false))
	require.Equal(t, "em ", f.ReflexivePrefix("1S", false, false))
	require.Equal(t, "me n'", f.ReflexivePrefix("1S", true, true))
	require.Equal(t, "me'n ", f.ReflexivePrefix("1S", false, true))
}
