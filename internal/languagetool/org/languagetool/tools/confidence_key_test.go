package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfidenceKey(t *testing.T) {
	k := NewConfidenceKey("en-US", "RULE")
	require.Equal(t, "en-US/RULE", k.String())
	require.True(t, k.Equal(NewConfidenceKey("en-US", "RULE")))
	require.False(t, k.Equal(NewConfidenceKey("de", "RULE")))
}
