package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchMatch_MAfter(t *testing.T) {
	m := NewSearchMatch("не був")
	tokens := []string{"Він", "не", "був", "тут"}
	require.Equal(t, 1, m.MAfter(tokens, 0))
	require.Equal(t, -1, m.MAfter(tokens, 2))
}

func TestSearchMatch_MBefore(t *testing.T) {
	m := NewSearchMatch("не був")
	tokens := []string{"Він", "не", "був", "тут"}
	require.Equal(t, 1, m.MBefore(tokens, 2))
}
