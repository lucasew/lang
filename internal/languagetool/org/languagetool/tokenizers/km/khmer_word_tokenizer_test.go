package km

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKhmerWordTokenizer(t *testing.T) {
	w := NewKhmerWordTokenizer()
	// Khmer danda U+17D4
	got := w.Tokenize("abc\u17D4def")
	require.Contains(t, got, "\u17D4")
	require.Contains(t, got, "abc")
	require.Contains(t, got, "def")
}
