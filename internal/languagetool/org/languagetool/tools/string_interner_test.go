package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntern(t *testing.T) {
	a := Intern("hello")
	b := Intern("hello")
	require.Equal(t, a, b)
	// same underlying data via map (string equality)
	require.True(t, a == b)
}
