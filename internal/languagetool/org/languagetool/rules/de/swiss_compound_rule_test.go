package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Unit twin of SwissCompoundRule.SwissExpander.expandLine (not a separate Java test class).
func TestSwissExpandLine(t *testing.T) {
	// Java: line with ß → [original, ß→ss]; else singleton.
	require.Equal(t, []string{"foo"}, SwissExpandLine("foo"))
	require.Equal(t, []string{"Fuß", "Fuss"}, SwissExpandLine("Fuß"))
	require.Equal(t, []string{"süß-sauer*", "süss-sauer*"}, SwissExpandLine("süß-sauer*"))
	require.Equal(t, []string{"no-beta"}, SwissExpandLine("no-beta"))
}
