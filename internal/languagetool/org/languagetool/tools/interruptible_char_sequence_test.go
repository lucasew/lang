package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterruptibleCharSequence(t *testing.T) {
	s := NewInterruptibleCharSequence("abc", nil)
	require.Equal(t, 3, s.Len())
	require.Equal(t, byte('a'), s.CharAt(0))
	require.Equal(t, "bc", s.SubSequence(1, 3).String())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s2 := NewInterruptibleCharSequence("x", ctx)
	require.Panics(t, func() { _ = s2.String() })
}
