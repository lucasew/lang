package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiKeyProperties(t *testing.T) {
	p := LoadMultiKeyProperties(strings.NewReader(`# c
a = 1
a=2
b = x
badline
`))
	require.Equal(t, []string{"1", "2"}, p.GetProperty("a"))
	require.Equal(t, []string{"x"}, p.GetProperty("b"))
	require.Nil(t, p.GetProperty("missing"))
}

// Twin: line.trim() is String.trim; split("\\s*=\\s*") is ASCII WS only.
func TestMultiKeyProperties_JavaTrimAndEqSplit(t *testing.T) {
	p := LoadMultiKeyProperties(strings.NewReader("  k\t=\tv  \n"))
	require.Equal(t, []string{"v"}, p.GetProperty("k"))
	// NBSP around = is not \\s without UNICODE_CHARACTER_CLASS → not a valid 2-part split
	p2 := LoadMultiKeyProperties(strings.NewReader("a\u00a0=\u00a0b\n"))
	require.Nil(t, p2.GetProperty("a"))
}
