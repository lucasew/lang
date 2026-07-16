package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSwissExpandLine(t *testing.T) {
	require.Equal(t, []string{"foo"}, SwissExpandLine("foo"))
	require.Equal(t, []string{"Fuß", "Fuss"}, SwissExpandLine("Fuß"))
}
