package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsWordCount_UTF16FirstUnit(t *testing.T) {
	// Java: length() > 0 && !isNotWordCharacter(substring(0,1)) — UTF-16 units
	require.True(t, isWordCount("Haus"))
	require.False(t, isWordCount(""))
	require.False(t, isWordCount("."))
	// emoji is 2 UTF-16 units; first unit is high surrogate → not a word character
	require.False(t, isWordCount("😀"))
}
