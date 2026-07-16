package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileLineExpander(t *testing.T) {
	e := FileLineExpanderFunc(func(line string) []string {
		return []string{line, line + "x"}
	})
	require.Equal(t, []string{"a", "ax"}, e.ExpandLine("a"))
	fn := AsLineExpander(e)
	require.Equal(t, []string{"b", "bx"}, fn("b"))
}
