package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemExitHandler(t *testing.T) {
	var got int
	prev := ExitHandler
	t.Cleanup(func() { ExitHandler = prev })
	SetSystemExitHandler(func(code int) { got = code })
	ExitHandler(7)
	require.Equal(t, 7, got)
}
