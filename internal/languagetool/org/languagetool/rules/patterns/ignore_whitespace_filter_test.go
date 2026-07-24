package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIgnoreWhitespaceFilter(t *testing.T) {
	f := IgnoreWhitespaceFilter{}
	require.Equal(t, FilterReject, f.AcceptText("  \n\t"))
	require.Equal(t, FilterReject, f.AcceptText(""))
	require.Equal(t, FilterAccept, f.AcceptText("word"))
	require.Equal(t, FilterAccept, f.AcceptElement("rule"))
	require.Equal(t, []string{"a", "b"}, FilterWhitespaceNodes([]string{"", "  ", "a", "\n", "b"}))
}
