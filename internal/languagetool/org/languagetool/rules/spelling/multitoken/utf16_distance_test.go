package multitoken

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Apache LevenshteinDistance uses CharSequence.charAt → UTF-16 units.
func TestRawLevenshtein_UTF16(t *testing.T) {
	// café vs cafe: differ by one BMP unit (é vs e)
	require.Equal(t, 1, rawLevenshtein("cafe", "café"))
	// emoji length 2 vs empty → distance 2 (not 1 code point)
	require.Equal(t, 2, rawLevenshtein("", "😀"))
	require.Equal(t, 0, rawLevenshtein("😀", "😀"))
}
