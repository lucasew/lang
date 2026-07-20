package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJwtContent(t *testing.T) {
	require.False(t, JwtNone.IsValid)
	require.False(t, JwtNone.IsPremium)
	require.Empty(t, JwtNone.Claims)
	c := NewJwtContent(true, true, map[string]any{"sub": "u1"})
	require.True(t, c.IsValid)
	require.True(t, c.IsPremium)
	require.Equal(t, "u1", c.Claims["sub"])
	// nil claims → empty map
	c2 := NewJwtContent(false, false, nil)
	require.NotNil(t, c2.Claims)
}
