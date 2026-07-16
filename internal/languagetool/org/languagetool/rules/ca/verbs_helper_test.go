package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsVerbDicendi(t *testing.T) {
	require.True(t, IsVerbDicendi("dir"))
	require.True(t, IsVerbDicendi("explicar"))
	require.False(t, IsVerbDicendi("correr"))
}

func TestIsVerbDicendiBefore(t *testing.T) {
	lemmas := []string{"SENT", "ell", "dir", "que"}
	// indices 1..3: ell, dir, que — keep looking on all for simplicity
	ok := IsVerbDicendiBefore(lemmas, 3, func(i int) bool { return i > 0 })
	require.True(t, ok)
	ok = IsVerbDicendiBefore(lemmas, 1, func(i int) bool { return i > 0 })
	require.False(t, ok) // only "ell" before stop
}
