package de

// Twin of GermanCompoundTokenizerTest (Java @Ignore interactive; green dict split)
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanCompoundTokenizer_Test(t *testing.T) {
	tok := NewGermanCompoundTokenizer(true)
	tok.AddWord("auto")
	tok.AddWord("bahn")
	tok.AddWord("haus")
	tok.AddWord("tür")
	// Capitalized compounds keep first-part capital (Java/jWordSplitter-ish surface)
	require.Equal(t, []string{"Auto", "bahn"}, tok.Tokenize("Autobahn"))
	require.Equal(t, []string{"Haus", "tür"}, tok.Tokenize("Haustür"))
	// unknown / short stays whole
	require.Equal(t, []string{"xyz"}, tok.Tokenize("xyz"))
	require.Equal(t, []string{"xyzabc"}, tok.Tokenize("xyzabc"))
}
