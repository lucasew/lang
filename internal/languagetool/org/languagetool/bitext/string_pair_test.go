package bitext

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.bitext.StringPair (no upstream unit test; behavior from source).

func TestStringPair(t *testing.T) {
	p := NewStringPair("src", "tgt")
	require.Equal(t, "src", p.GetSource())
	require.Equal(t, "tgt", p.GetTarget())
	// Java toString: sourceString + " & " + targetString
	require.Equal(t, "src & tgt", p.String())

	empty := NewStringPair("", "")
	require.Equal(t, " & ", empty.String())
}
