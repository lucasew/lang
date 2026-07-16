package nl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompoundAcceptor(t *testing.T) {
	c := NewCompoundAcceptor()
	require.NoError(t, c.LoadNoS(strings.NewReader("auto\n")))
	c.KnownWords["auto"] = struct{}{}
	c.KnownWords["weg"] = struct{}{}
	require.True(t, c.Accept("autoweg"))
	require.True(t, c.Accept("TV-show"))
	require.False(t, c.Accept("xyz"))
}
