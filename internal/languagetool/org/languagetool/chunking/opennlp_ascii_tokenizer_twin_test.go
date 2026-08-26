package chunking

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin: java.util.StringTokenizer default delimiters " \t\n\r\f".
func TestASCIIStringTokenizerSplit(t *testing.T) {
	require.Equal(t, []string{"1", "2", "3"}, asciiStringTokenizerSplit("1 2\t3"))
	require.Equal(t, []string{"1", "2"}, asciiStringTokenizerSplit("  1   2  "))
	// NBSP is not a default StringTokenizer delimiter
	require.Equal(t, []string{"1\u00a02"}, asciiStringTokenizerSplit("1\u00a02"))
}
