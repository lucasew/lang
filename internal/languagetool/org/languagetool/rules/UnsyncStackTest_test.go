package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.UnsyncStackTest

func TestUnsyncStack_Stack(t *testing.T) {
	stack := NewUnsyncStack[string]()
	require.True(t, stack.Empty())
	stack.Push("test")
	require.Equal(t, "test", stack.Peek())
	require.False(t, stack.Empty())
	require.Equal(t, "test", stack.Pop())
	require.True(t, stack.Empty())
	require.Panics(t, func() { stack.Pop() })
	require.Panics(t, func() { stack.Peek() })
}
