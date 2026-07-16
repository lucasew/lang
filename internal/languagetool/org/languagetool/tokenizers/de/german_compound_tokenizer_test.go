package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanCompoundTokenizer(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	tok.AddWord("auto")
	tok.AddWord("bahn")
	got := tok.Tokenize("autobahn")
	require.Equal(t, []string{"auto", "bahn"}, got)
	// unknown stays whole
	require.Equal(t, []string{"xyzabc"}, tok.Tokenize("xyzabc"))
}
