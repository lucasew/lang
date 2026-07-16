package ml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMalayalamWordTokenizer(t *testing.T) {
	w := NewMalayalamWordTokenizer()
	got := w.Tokenize("hello, world")
	require.Contains(t, got, ",")
	require.Contains(t, got, "hello")
	require.Contains(t, got, "world")
}
