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
